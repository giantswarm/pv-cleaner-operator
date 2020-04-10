package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"k8s.io/client-go/rest"

	"github.com/giantswarm/pv-cleaner-operator/flag"
	"github.com/giantswarm/pv-cleaner-operator/service/controller"
)

type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper

	Description string
	GitCommit   string
	ProjectName string
	Source      string
	Version     string
}

type Service struct {
	Version *version.Service

	bootOnce                   sync.Once
	persistentVolumeController *controller.PersistentVolume
}

func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	config.Logger.Log("debug", fmt.Sprintf("creating pv-cleaner-operator gitCommit:%s", config.GitCommit))

	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:    config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster:  config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			KubeConfig: config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig),
			TLS: k8srestconfig.ConfigTLS{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			Logger:     config.Logger,
			RestConfig: restConfig,
		}
		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var persistentVolumeController *controller.PersistentVolume
	{
		c := controller.PersistentVolumeConfig{
			K8sClient: k8sClient,
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		persistentVolumeController, err = controller.NewPersistentVolume(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	}

	var versionService *version.Service
	{
		c := version.Config{
			Description: config.Description,
			GitCommit:   config.GitCommit,
			Name:        config.ProjectName,
			Source:      config.Source,
			Version:     config.Version,
		}

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Version: versionService,

		bootOnce:                   sync.Once{},
		persistentVolumeController: persistentVolumeController,
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.persistentVolumeController.Boot(context.Background())
	})
}
