package controller

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/giantswarm/pv-cleaner-operator/service/controller/v1"
)

const (
	cleanupLabel = "persistentvolume.giantswarm.io/cleanup-on-release"
)

type PersistentVolumeConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	ProjectName string
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

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.ProjectName must not be empty")
	}

	var v1ResourceSet *controller.ResourceSet
	{
		c := v1.ResourceSetConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ProjectName: config.ProjectName,
		}

		v1ResourceSet, err = v1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var persistentVolumeController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(corev1.PersistentVolume)
			},
			Logger: config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v1ResourceSet,
			},
			Selector: labels.SelectorFromSet(map[string]string{
				cleanupLabel: "true",
			}),

			Name: config.ProjectName,
		}

		persistentVolumeController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	p := &PersistentVolume{
		Controller: persistentVolumeController,
	}

	return p, nil
}
