package persistentvolume

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	cleanupPV, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetUpdateChange(cleanupPV)

	return patch, nil
}

// make some actions on changing resource
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateState interface{}) error {
	pv, err := toPV(updateState)
	if err != nil {
		return microerror.Mask(err)
	}

	if pv.Name == "" {
		// Nothing to do.
		return nil
	}

	r.logger.Log("pv", pv.GetName(), "debug", "looking for annotations on pv")
	recycleStateAnnotationValue := getRecycleStateAnnotation(pv, RecycleStateAnnotation)

	if recycleStateAnnotationValue == Recycled && pv.Status.Phase == "Released" {
		r.logger.Log("pv", pv.GetName(), "debug", "update recycle state annotation")
		pv, err = getUpdatedRecycleStateAnnotation(pv)
		_, err := r.k8sClient.Core().PersistentVolumes().Update(pv)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if (recycleStateAnnotationValue == Released || recycleStateAnnotationValue == Cleaning) && pv.Status.Phase == "Released" {
		r.logger.Log("pv", pv.GetName(), "debug", "delete released pv")
		err = r.k8sClient.Core().PersistentVolumes().Delete(pv.Name, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if recycleStateAnnotationValue == Cleaning && pv.Status.Phase != "Released" {
		pvcdef := getPvc(pv)
		pvc, err := r.k8sClient.Core().PersistentVolumeClaims("kube-system").Get(pvcdef.GetName(), metav1.GetOptions{})
		if err != nil {
			pvc, err = r.k8sClient.Core().PersistentVolumeClaims("kube-system").Create(pvcdef)
			if err != nil {
				return microerror.Maskf(err, "Failed to create pvc", pvc.GetName())
			}
		}

		if pvc.Status.Phase != "Bound" {
			r.logger.Log("pvc", pvc.GetName(), "debug", "waiting for pvc to bound pv")
			return nil
		}

		cleanupJobDef := getCleanupJob(pvc)
		cleanupJob, err := r.k8sClient.Batch().Jobs("kube-system").Create(cleanupJobDef)
		if err != nil {
			cleanupJob, err = r.k8sClient.Batch().Jobs("kube-system").Get(cleanupJobDef.GetName(), metav1.GetOptions{})
			if err != nil {
				return microerror.Maskf(err, "Failed to create pvc", pvc.GetName())
			}
		}

		if cleanupJob.Status.Succeeded != 1 {
			r.logger.Log("job", cleanupJob.GetName(), "pv", pv.GetName(), "debug", "Waiting for job to complete cleanup of pv")
			return nil
		}

		gracePeriodSeconds := int64(0)
		deleteOptions := &metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		}
		if err := r.k8sClient.Batch().Jobs("kube-system").Delete(cleanupJob.Name, deleteOptions); err != nil {
			return microerror.Maskf(err, "Failed to delete recycle job", cleanupJob.GetName())
		}

		if err := r.k8sClient.Core().PersistentVolumeClaims("kube-system").Delete(pvc.Name, deleteOptions); err != nil {
			return microerror.Maskf(err, "Failed to delete recycle pvc", pvc.GetName())
		}
	}

	return nil
}

// get objects which should be somehow processed e.g. have annotation
func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentPV, err := toPV(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if !isScheduledForCleanup(currentPV, CleanupAnnotation) {
		return &apiv1.PersistentVolume{}, nil
	}

	return currentPV, nil
}
