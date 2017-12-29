package endpoint

import (
	"github.com/giantswarm/microendpoint/endpoint/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/pv-cleaner-operator/service"
)

type Config struct {
	Logger  micrologger.Logger
	Service *service.Service
}

func DefaultConfig() Config {
	return Config{
		Logger:  nil,
		Service: nil,
	}
}

type Endpoint struct {
	Version *version.Endpoint
}

func New(config Config) (*Endpoint, error) {
	var err error

	var newVersionEndpoint *version.Endpoint
	{
		versionConfig := version.DefaultConfig()

		versionConfig.Logger = config.Logger
		versionConfig.Service = config.Service.Version

		newVersionEndpoint, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newEndpoint := &Endpoint{
		Version: newVersionEndpoint,
	}

	return newEndpoint, nil
}
