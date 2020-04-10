package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
)

// GetCurrentState returns current state of the recycled persistent volume.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {

	pv, err := toPV(obj)
	if err != nil {
		return nil, microerror.Mask(err)	}

	recycleState := getVolumeAnnotation(pv, recycleStateAnnotation)
	rpv := &RecyclePersistentVolume{
		Name:         pv.Name,
		State:        pv.Status.Phase,
		RecycleState: recycleState,
	}

	return rpv, nil
}

// GetDesiredState returns desired state of the recycled persistent volume.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {

	pv, err := toPV(obj)
	if err != nil {
		return nil, microerror.Mask(err)	}

	rpv := &RecyclePersistentVolume{
		Name:         pv.Name,
		State:        "Available",
		RecycleState: recycled,
	}

	return rpv, nil
}
