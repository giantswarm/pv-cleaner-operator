package operator

import (
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
	"k8s.io/client-go/kubernetes"
)

const ResyncPeriod = 1 * time.Minute

func newFramework(config Config) (*framework.Framework, error) {

	var k8sClient kubernetes.Interface
	var err error

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()
		c.ResyncPeriod = ResyncPeriod
		c.Watcher = k8sClient.Core().PersistentVolumes()

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var f *framework.Framework
	{
		resources := []framework.Resource{}

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
