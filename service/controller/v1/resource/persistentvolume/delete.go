package persistentvolume

import (
	"context"

	"github.com/giantswarm/operatorkit/controller"
)

// NewDeletePatch returns patch to apply on deleted persistent volume.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}

// ApplyDeleteChange represents delete patch logic.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteState interface{}) error {
	return nil
}
