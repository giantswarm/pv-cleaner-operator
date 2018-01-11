package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

// NewDeletePatch returns patch to apply on deleted persistent volume.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	deleteState, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(deleteState)

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

	oldpv, err := toPV(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var recycleStateAnnotationValue string
	switch rpv.RecycleState {
	case "Released":
		recycleStateAnnotationValue = cleaning
	case "Cleaning":
		recycleStateAnnotationValue = recycled
	}

	newpv, err := r.newRecycleStateAnnotation(oldpv, recycleStateAnnotationValue)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("pv", newpv.Name, "create volume with new recycle annotation", recycleStateAnnotationValue)
	_, err = r.k8sClient.Core().PersistentVolumes().Create(newpv)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// newDeleteChange checks wherether persistent volume should be reconciled
// on delete event.
func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	deletedVolume, err := toPV(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "persistentvolume", deletedVolume.Name, "retrieving cleanup annotation of deleted volume", cleanupAnnotation)
	reconcile := isScheduledForCleanup(deletedVolume, cleanupAnnotation)
	r.logger.LogCtx(ctx, "persistentvolume", deletedVolume.Name, "reconcile deleted persistent volume", reconcile)
	if !reconcile {
		return nil, nil
	}

	return currentState, nil
}
