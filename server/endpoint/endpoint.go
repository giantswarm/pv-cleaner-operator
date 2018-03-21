package endpoint

import (
	versionendpoint "github.com/giantswarm/microendpoint/endpoint/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/pv-cleaner-operator/service"
)

// Config represents the configuration used to create a endpoint.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Service *service.Service
}

// New creates a new configured endpoint.
func New(config Config) (*Endpoint, error) {
	var err error

	var versionEndpoint *versionendpoint.Endpoint
	{
		versionConfig := versionendpoint.Config{}
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
