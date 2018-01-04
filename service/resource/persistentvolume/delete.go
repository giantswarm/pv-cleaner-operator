package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/api/core/v1"
)

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteState interface{}) error {
	oldpv, err := toPV(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if oldpv.Name == "" {
		// Nothing to do.
		return nil
	}

	r.logger.Log("pv", oldpv.GetName(), "debug", "looking for annotations on deleted pv")
	recycleStateAnnotationValue := getRecycleStateAnnotation(oldpv, RecycleStateAnnotation)

	if recycleStateAnnotationValue == Released || recycleStateAnnotationValue == Cleaning {
		newpv, err := getUpdatedRecycleStateAnnotation(oldpv)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Log("pv", newpv.GetName(), "debug", "create volume")
		_, err = r.k8sClient.Core().PersistentVolumes().Create(newpv)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	deletedPV, err := toPV(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if !isScheduledForCleanup(deletedPV, CleanupAnnotation) {
		return &apiv1.PersistentVolume{}, nil
	}

	return deletedPV, nil
}
