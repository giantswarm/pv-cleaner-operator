package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
)

// GetCurrentState returns current state of the recycled persistent volume
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {

	rpv, err := pvToRecyclePV(obj)
	if err != nil {
		return nil, microerror.Maskf(err, "GetCurrentState")
	}

	return rpv, nil
}

// GetDesiredState returns desired state of the recycled persistent volume
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {

	pv, err := pvToRecyclePV(obj)
	if err != nil {
		return nil, microerror.Maskf(err, "GetDesiredState")
	}

	rpv := &RecyclePersistentVolume{
		Name:         pv.Name,
		State:        "Available",
		RecycleState: recycled,
	}
	return rpv, nil
}
