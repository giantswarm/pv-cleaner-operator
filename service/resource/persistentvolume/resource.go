package persistentvolume

import (
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/api/core/v1"
)

const (
	name                   = "persistentvolume"
	cleanupAnnotation      = "persistentvolume.giantswarm.io/cleanup"
	recycleStateAnnotation = "persistentvolume.giantswarm.io/recyclestate"
)

// RecycleStateAnnotation values
const (
	released string = "Released"
	cleaning string = "Cleaning"
	recycled string = "Recycled"
)

// Config describes resource configuration
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

// Resource stores resource configuration
type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New is factory for resource objects
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

// Name returns name of the managed resource
func (r *Resource) Name() string {
	return name
}

// Underlying returns managed resource object
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
		recycleStateAnnotationValue = recycled
	}
	return recycleStateAnnotationValue
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
