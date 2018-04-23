package controller

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/pv-cleaner-operator/service/controller/v1"
)

const cleanupLabel = "persistentvolume.giantswarm.io/cleanup-on-release"

type PersistentVolumeConfig struct {
	K8sClient kubernetes.Interface
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

	var newInformer *informer.Informer
	{
		c := informer.Config{
			ResyncPeriod: informer.DefaultResyncPeriod,
			Watcher:      config.K8sClient.Core().PersistentVolumes(),
			ListOptions: metav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=%s", cleanupLabel, "true"),
			},
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

			ProjectName: config.ProjectName,
		}

		v1ResourceSet, err = v1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *controller.ResourceRouter
	{
		c := controller.ResourceRouterConfig{
			Logger: config.Logger,

			ResourceSets: []*controller.ResourceSet{
				v1ResourceSet,
			},
		}

		resourceRouter, err = controller.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var f *controller.Controller
	{
		c := controller.Config{
			Informer:       newInformer,
			Logger:         config.Logger,
			ResourceRouter: resourceRouter,
			RESTClient:     config.K8sClient.CoreV1().RESTClient(),

			Name: config.ProjectName,
		}

		f, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	p := &PersistentVolume{
		Controller: f,
	}

	return p, nil
}
