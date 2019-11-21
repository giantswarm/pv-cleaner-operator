package persistentvolume

import (
	"context"

	"github.com/giantswarm/operatorkit/resource/crud"
)

// NewCreatePatch is not used during persistent volume reconcile.
func (r *Resource) NewCreatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	return nil, nil
}

// ApplyCreateChange is not used during persistent volume reconcile.
func (r *Resource) ApplyCreateChange(ctx context.Context, obj, deleteState interface{}) error {
	return nil
}
