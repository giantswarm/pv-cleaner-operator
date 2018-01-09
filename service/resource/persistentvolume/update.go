package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

// NewUpdatePatch returns patch to apply on deleted persistent volume
func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetUpdateChange(update)

	return patch, nil
}

// ApplyUpdateChange represents delete patch logic
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateState interface{}) error {
	return nil
}

// newUpdateChange checks wherether persistent volume should be reconciled
// on update event
func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	return nil, nil
}
