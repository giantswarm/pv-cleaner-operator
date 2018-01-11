package persistentvolume

import (
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	name                   = "persistentvolume"
	cleanupAnnotation      = "volume.kubernetes.io/cleanup-on-release"
	recycleStateAnnotation = "pv-cleaner-operator.giantswarm.io/volume-recycle-state"
)

const (
	released string = "Released"
	cleaning string = "Cleaning"
	recycled string = "Recycled"
)

// Config describes resource configuration.
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

// Underlying returns managed resource object.
func (r *Resource) Underlying() framework.Resource {
	return r
}

// isScheduledForCleanup checks whethere persistent volume object has cleanup annotation.
func isScheduledForCleanup(pv *apiv1.PersistentVolume, cleanupAnnotation string) bool {
	cleanupAnnotationValue, ok := pv.Annotations[cleanupAnnotation]
	return ok && cleanupAnnotationValue == "true"
}

// getRecycleStateAnnotation returns current recycle state annotation.
// If it is empty - 'recycled' annotation returned,
// so that volumes, which were never recycled before by the operator
// will be considered in the same way as recycled volumes.
func getRecycleStateAnnotation(pv *apiv1.PersistentVolume, recycleStateAnnotation string) (recycleStateAnnotationValue string) {
	recycleStateAnnotationValue, ok := pv.Annotations[recycleStateAnnotation]
	if !ok {
		recycleStateAnnotationValue = recycled
	}
	return recycleStateAnnotationValue
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
		RecycleState: getRecycleStateAnnotation(pv, recycleStateAnnotation),
	}

	return rpv, nil
}

func (r *Resource) newRecycleStateAnnotation(pv *apiv1.PersistentVolume, recycleAnnotation string) (*apiv1.PersistentVolume, error) {

	updatedpv := &apiv1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        pv.Name,
			Annotations: pv.Annotations,
			Labels: pv.Labels,
		},
		Spec: apiv1.PersistentVolumeSpec{
			Capacity:                      pv.Spec.Capacity,
			StorageClassName:              pv.Spec.StorageClassName,
			AccessModes:                   pv.Spec.AccessModes,
			PersistentVolumeReclaimPolicy: pv.Spec.PersistentVolumeReclaimPolicy,
			PersistentVolumeSource:        pv.Spec.PersistentVolumeSource,
		},
	}

	r.logger.Log("persistentvolume", pv.Name, "set new recycle annotation", recycleAnnotation)
	updatedpv.ObjectMeta.Annotations[recycleStateAnnotation] = recycleAnnotation
	
	return updatedpv, nil
}
