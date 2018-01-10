package persistentvolume

import (
	"context"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

// NewDeletePatch returns patch to apply on deleted persistent volume.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	deleteState, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetUpdateChange(deleteState)

	return patch, nil
}

// ApplyDeleteChange represents delete patch logic.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteState interface{}) error {
	rpv, err := toRecyclePV(deleteState)
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

	// delete state logic

	return nil
}

// newDeleteChange checks wherether persistent volume should be reconciled
// on delete event.
func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	deletedVolume, err := toPV(obj)
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
