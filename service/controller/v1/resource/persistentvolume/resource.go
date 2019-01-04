package persistentvolume

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultStorageClass    = "default"
	name                   = "persistentvolume"
	storageClassAnnotation = "volume.beta.kubernetes.io/storage-class"
	recycleStateAnnotation = "pv-cleaner-operator.giantswarm.io/volume-recycle-state"
)

const (
	cleaning string = "Cleaning"
	recycled string = "Recycled"
)

// Config describes resource configuration.
type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// Resource stores resource configuration.
type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New is factory for resource objects.
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	resource := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}
	return resource, nil
}

// Name returns name of the managed resource.
func (r *Resource) Name() string {
	return name
}

// getVolumeAnnotation returns current recycle state annotation.
// If it is empty - 'recycled' annotation returned,
// so that volumes, which were never recycled before by the operator
// will be considered in the same way as recycled volumes.
func getVolumeAnnotation(pv *apiv1.PersistentVolume, annotation string) (annotationValue string) {
	annotationValue, ok := pv.Annotations[annotation]
	if !ok {
		annotationValue = recycled
	}
	return annotationValue
}

// toPV converts interface object into PersistentVolume object.
func toPV(v interface{}) (*apiv1.PersistentVolume, error) {
	if v == nil {
		return nil, nil
	}

	pv, ok := v.(*apiv1.PersistentVolume)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1.PersistentVolume{}, v)
	}

	return pv, nil
}

// toRecyclePV converts interface object into RecyclePersistentVolume object.
func toRecyclePV(v interface{}) (*RecyclePersistentVolume, error) {
	if v == nil {
		return nil, nil
	}

	pv, ok := v.(*RecyclePersistentVolume)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &RecyclePersistentVolume{}, v)
	}

	return pv, nil
}

// pvToRecyclePV creates RecyclePersistentVolume object from PersistentVolume object.
func pvToRecyclePV(v interface{}) (*RecyclePersistentVolume, error) {
	if v == nil {
		return nil, nil
	}

	pv, err := toPV(v)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	rpv := &RecyclePersistentVolume{
		Name:         pv.Name,
		State:        pv.Status.Phase,
		RecycleState: getVolumeAnnotation(pv, recycleStateAnnotation),
	}

	return rpv, nil
}

// newRecycleStateAnnotation create new PersistentVolume object with
// updated recycle state annotation.
func (r *Resource) newRecycleStateAnnotation(pv *apiv1.PersistentVolume, recycleAnnotation string) (*apiv1.PersistentVolume, error) {

	updatedpv := &apiv1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        pv.Name,
			Annotations: pv.Annotations,
			Labels:      pv.Labels,
			Finalizers:  []string{},
		},
		Spec: apiv1.PersistentVolumeSpec{
			Capacity:                      pv.Spec.Capacity,
			AccessModes:                   pv.Spec.AccessModes,
			StorageClassName:              pv.Spec.StorageClassName,
			PersistentVolumeReclaimPolicy: pv.Spec.PersistentVolumeReclaimPolicy,
			PersistentVolumeSource:        pv.Spec.PersistentVolumeSource,
		},
	}

	r.logger.Log("persistentvolume", pv.Name, "set new recycle annotation", recycleAnnotation)
	updatedpv.ObjectMeta.Annotations[recycleStateAnnotation] = recycleAnnotation

	return updatedpv, nil
}

// newPvc returns k8s PersistentVolumeClaim object,
// which bounds persistent volume from function parameter.
func newPvc(pv *apiv1.PersistentVolume) *apiv1.PersistentVolumeClaim {

	storageClassAnnotationValue, ok := pv.Annotations[storageClassAnnotation]
	if !ok {
		if pv.Spec.StorageClassName != "" {
			storageClassAnnotationValue = pv.Spec.StorageClassName
		} else {
			storageClassAnnotationValue = defaultStorageClass
		}
	}

	pvc := &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:       fmt.Sprintf("pv-cleaner-claim-%s", pv.Name),
			Namespace:  "kube-system",
			Finalizers: []string{},
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes:      pv.Spec.AccessModes,
			StorageClassName: &storageClassAnnotationValue,
			Resources: apiv1.ResourceRequirements{
				Requests: pv.Spec.Capacity,
			},
			VolumeName: pv.Name,
		},
	}

	return pvc
}

// newCleanupJob returns k8s job objects,
// which runs busybox container, mounts claim from the function parameter
// and run shell command to cleanup mount path.
func newCleanupJob(pvc *apiv1.PersistentVolumeClaim) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("pv-cleaner-job-%s", pvc.Name),
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pv-cleaner-pod",
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						apiv1.Container{
							Name:  fmt.Sprintf("pv-cleaner-container-%s", pvc.Name),
							Image: "busybox",
							Command: []string{
								"/bin/sh",
								"-c",
								"test -e /scrub && rm -rf /scrub/..?* /scrub/.[!.]* /scrub/*  && test -z \"$(ls -A /scrub)\" || exit 1",
							},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "pv-cleaner-mount",
									MountPath: "/scrub",
								},
							},
						},
					},
					RestartPolicy: "Never",
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "pv-cleaner-mount",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvc.Name,
								},
							},
						},
					},
				},
			},
		},
	}

	return job
}
