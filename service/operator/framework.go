package operator

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/pv-cleaner-operator/service/resource/persistentvolume"
)

func newFramework(config Config) (*framework.Framework, error) {

	c := persistentvolume.DefaultConfig()
	c.K8sClient = config.K8sClient
	c.Logger = config.Logger

	var persistentVolumeResource, err = persistentvolume.New(c)
	if err != nil {
		return nil, microerror.Maskf(err, "persistentvolume.New")
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()
		c.Watcher = config.K8sClient.Core().PersistentVolumes()

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var f *framework.Framework
	{
		resources := []framework.Resource{
			persistentVolumeResource,
		}

		c := framework.DefaultConfig()

		c.Logger = config.Logger
		c.ResourceRouter = framework.DefaultResourceRouter(resources)
		c.Informer = newInformer

		f, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return f, nil
}
