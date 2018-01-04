package persistentvolume

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
)

const (
	Name                   = "persistentvolume"
	CleanupAnnotation      = "persistentvolume.giantswarm.io/cleanup"
	RecycleStateAnnotation = "persistentvolume.giantswarm.io/recyclestate"
)

// RecycleStateAnnotation values
const (
	Released string = "Released"
	Cleaning string = "Cleaning"
	Recycled string = "Recycled"
)

type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new resource by
// best effort.
func DefaultConfig() Config {
	return Config{
		K8sClient: nil,
		Logger:    nil,
	}
}

type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

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

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func isScheduledForCleanup(pv *apiv1.PersistentVolume, cleanupAnnotation string) bool {
	cleanupAnnotationValue, ok := pv.Annotations[cleanupAnnotation]
	return ok && cleanupAnnotationValue == "true"
}

func getRecycleStateAnnotation(pv *apiv1.PersistentVolume, recycleStateAnnotation string) (recycleStateAnnotationValue string) {
	recycleStateAnnotationValue, ok := pv.Annotations[recycleStateAnnotation]
	if !ok {
		recycleStateAnnotationValue = Recycled
	}
	return recycleStateAnnotationValue
}

func getUpdatedRecycleStateAnnotation(pv *apiv1.PersistentVolume) (*apiv1.PersistentVolume, error) {

	recycleStateAnnotationValue := getRecycleStateAnnotation(pv, RecycleStateAnnotation)

	switch recycleStateAnnotationValue {
	case Released:
		recycleStateAnnotationValue = Cleaning
	case Cleaning:
		recycleStateAnnotationValue = Recycled
	case Recycled:
		recycleStateAnnotationValue = Released
	}

	var updatedpv *apiv1.PersistentVolume
	{
		allAnnotations := pv.GetAnnotations()
		allAnnotations[RecycleStateAnnotation] = recycleStateAnnotationValue
		updatedpv = &apiv1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:        pv.Name,
				Annotations: allAnnotations,
			},
			Spec: apiv1.PersistentVolumeSpec{
				Capacity:                      pv.Spec.Capacity,
				StorageClassName:              pv.Spec.StorageClassName,
				AccessModes:                   pv.Spec.AccessModes,
				PersistentVolumeReclaimPolicy: pv.Spec.PersistentVolumeReclaimPolicy,
				PersistentVolumeSource:        pv.Spec.PersistentVolumeSource,
			},
		}
	}

	return updatedpv, nil
}

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

func getPvc(pv *apiv1.PersistentVolume) *apiv1.PersistentVolumeClaim {

	pvc := &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pv-cleaner-claim-%s", pv.GetName()),
			Namespace: "kube-system",
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			StorageClassName: &pv.Spec.StorageClassName,
			AccessModes:      pv.Spec.AccessModes,
			Resources: apiv1.ResourceRequirements{
				Requests: pv.Spec.Capacity,
			},
			VolumeName: pv.GetName(),
		},
	}

	return pvc
}

func getCleanupJob(pvc *apiv1.PersistentVolumeClaim) *batchv1.Job {

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("pv-cleaner-job-%s", pvc.GetName()),
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pv-cleaner-pod",
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						apiv1.Container{
							Name:  fmt.Sprintf("pv-cleaner-container-%s", pvc.GetName()),
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
