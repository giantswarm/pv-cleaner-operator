package persistentvolume

import (
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/client-go/pkg/api/v1"
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

func isScheduledForCleanup(pv apiv1.PersistentVolume, cleanupAnnotation string) bool {
	cleanupAnnotationValue, ok := pv.GetAnnotations()[cleanupAnnotation]
	return ok && cleanupAnnotationValue == "true"
}

func getRecycleStateAnnotation(pv apiv1.PersistentVolume, recycleStateAnnotation string) (recycleStateAnnotationValue string) {
	recycleStateAnnotationValue, ok := pv.GetAnnotations()[recycleStateAnnotation]
	if !ok {
		recycleStateAnnotationValue = ""
	}
	return recycleStateAnnotationValue
}

func updateRecycleStateAnnotation(oldstate string) string {
	var newstate string

	switch oldstate {
	case Released:
		newstate = Cleaning
	case Cleaning:
		newstate = Recycled
	case Recycled:
		newstate = Released
	}

	return newstate
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
