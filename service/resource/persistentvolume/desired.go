package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	pv, err := toPV(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredPersistentVolume := PersistentVolume{
		Name: pv.Name,
		State: "Available",
		RecycleState: Recycled,
	}
	
	return desiredPersistentVolume, nil
}
