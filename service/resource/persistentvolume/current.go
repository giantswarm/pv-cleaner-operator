package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	pv, err := toPV(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("pv", pv.GetName(), "debug", "looking for annotations on pv")

	recycleStateAnnotation := getRecycleStateAnnotation(*pv, RecycleStateAnnotation)

	currentPersistentVolume := PersistentVolume{
		Name:         pv.Name,
		State:        string(pv.Status.Phase),
		RecycleState: recycleStateAnnotation,
	}

	return currentPersistentVolume, nil
}
