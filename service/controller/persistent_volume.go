package controller

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/pv-cleaner-operator/pkg/project"
	v1 "github.com/giantswarm/pv-cleaner-operator/service/controller/v1"
)

const (
	cleanupLabel = "persistentvolume.giantswarm.io/cleanup-on-release"
)

type PersistentVolumeConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type PersistentVolume struct {
	*controller.Controller
}

func NewPersistentVolume(config PersistentVolumeConfig) (*PersistentVolume, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	var v1ResourceSet *controller.ResourceSet
	{
		c := v1.ResourceSetConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		v1ResourceSet, err = v1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			Name:      project.Name(),
			ResourceSets: []*controller.ResourceSet{
				v1ResourceSet,
			},
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(corev1.PersistentVolume)
			},
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	p := &PersistentVolume{
		Controller: operatorkitController,
	}

	return p, nil
}
