package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
)

// NewDeletePatch returns patch to apply on deleted persistent volume.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {

	patch := controller.NewPatch()
	patch.SetDeleteChange(currentState)

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
	case "Recycled":
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
