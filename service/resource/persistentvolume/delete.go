package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

// NewDeletePatch returns patch to apply on deleted persistent volume
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

// ApplyDeleteChange represents delete patch logic
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteState interface{}) error {
	return nil
}

// newDeleteChange checks wherether persistent volume should be reconciled
// on delete event
func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	return nil, nil
}
