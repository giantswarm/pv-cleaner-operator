package persistentvolume

import (
	"context"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

// NewUpdatePatch returns patch to apply on deleted persistent volume.
func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	updateState, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetUpdateChange(updateState)

	return patch, nil
}

// ApplyUpdateChange represents delete patch logic.
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateState interface{}) error {
	rpv, err := toRecyclePV(updateState)
	if err != nil {
		return microerror.Mask(err)
	}

	if rpv == nil {
		// Nothing to do.
		return nil
	}

	_, err = toPV(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// update state logic

	return nil
}

// newUpdateChange checks wherether persistent volume should be reconciled
// on update event.
func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	updatedVolume, err := toPV(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "persistentvolume", updatedVolume.Name, "retrieving cleanup annotation", cleanupAnnotation)
	reconcile := isScheduledForCleanup(updatedVolume, cleanupAnnotation)
	r.logger.LogCtx(ctx, "persistentvolume", updatedVolume.Name, "reconcile persistent volume", reconcile)
	if !reconcile {
		return nil, nil
	}

	if reflect.DeepEqual(currentState, desiredState) {
		r.logger.LogCtx(ctx, "persistentvolume", updatedVolume.Name, "volume reconciled to desired state", "true")
		return nil, nil
	}

	return currentState, nil
}
