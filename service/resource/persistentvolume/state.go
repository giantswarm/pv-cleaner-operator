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

	return pv, nil
}

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	pv, err := toPV(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return pv, nil
}
