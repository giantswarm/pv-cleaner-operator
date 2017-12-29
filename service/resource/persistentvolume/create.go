package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) NewCreatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	delete, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(delete)

	return patch, nil
}

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, deleteState interface{}) error {
	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	return nil, nil
}
