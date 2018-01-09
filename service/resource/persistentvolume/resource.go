package persistentvolume

import (
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	name = "endpoint"
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
