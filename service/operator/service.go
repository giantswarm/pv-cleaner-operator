package operator

import (
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/client-go/kubernetes"
)

// Config represents the configuration used to create an Operator service.
type Config struct {
	// Dependencies.

	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new operator
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// Service implements the Operator service interface.
type Service struct {
	// Dependencies.

	logger micrologger.Logger

	// Internals.

	framework *framework.Framework
	bootOnce  sync.Once
}

// New creates a new configured Operator service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	operatorFramework, err := newFramework(config)
	if err != nil {
		return nil, microerror.Maskf(err, "newFramework")
	}

	newService := &Service{
		// Dependencies.
		logger: config.Logger,

		// Internals.
		framework: operatorFramework,
		bootOnce:  sync.Once{},
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		go s.framework.Boot()
	})
}
