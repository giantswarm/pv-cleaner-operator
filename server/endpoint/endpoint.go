package endpoint

import (
	versionendpoint "github.com/giantswarm/microendpoint/endpoint/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/pv-cleaner-operator/service"
)

// Config represents the configuration used to create a endpoint.
type Config struct {
	Logger  micrologger.Logger
	Service *service.Service
}

// DefaultConfig provides a default configuration to create a new endpoint by
// best effort.
func DefaultConfig() Config {
	return Config{
		Logger:  nil,
		Service: nil,
	}
}

func New(config Config) (*Endpoint, error) {
	var err error

	var versionEndpoint *versionendpoint.Endpoint
	{
		versionConfig := versionendpoint.DefaultConfig()
		versionConfig.Logger = config.Logger
		versionConfig.Service = config.Service.Version
		versionEndpoint, err = versionendpoint.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newEndpoint := &Endpoint{
		Version: versionEndpoint,
	}

	return newEndpoint, nil
}

// Endpoint is the endpoint collection.
type Endpoint struct {
	Version *versionendpoint.Endpoint
}
