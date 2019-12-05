package controller

import (
	"fmt"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Logger:  config.Logger,
			Watcher: config.K8sClient.K8sClient().CoreV1().PersistentVolumes(),

			ListOptions: metav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=%s", cleanupLabel, "true"),
			},
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
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
			Informer: newInformer,
			Logger:   config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v1ResourceSet,
			},
			RESTClient: config.K8sClient.RESTClient(),

			Name: project.Name(),
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
