package persistentvolume

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewUpdatePatch returns patch to apply on updated persistent volume.
func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {

	updateState, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetUpdateChange(updateState)

	return patch, nil
}

// ApplyUpdateChange represents update patch logic.
// All actions are based on combination of volume state
// and custom recycle state.
//   * ReleasedRecycled - initial state of volume after claim is deleted; volume is recreated at this step
//   * AvailableCleaning - volume ready for bounding to cleanup claim
//   * BoundCleaning - volume claim is ready for mounting into cleanup job
//   * ReleasedCleaning - volume claim was succesfully cleaned up, volume can be recreated
//   * AvailableRecycled - desired state of the volume
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateState interface{}) error {
	rpv, err := toRecyclePV(updateState)
	if err != nil {
		return microerror.Mask(err)
	}

	if rpv == nil {
		// Nothing to do.
		return nil
	}

	pv, err := toPV(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	switch combinedState := string(rpv.State) + rpv.RecycleState; combinedState {
	case "ReleasedRecycled":
		pv, err := r.newRecycleStateAnnotation(pv, recycled)
		_, err = r.k8sClient.Core().PersistentVolumes().Update(pv)
		if err != nil {
			return microerror.Mask(err)
		}
		err = r.k8sClient.Core().PersistentVolumes().Delete(pv.Name, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	case "AvailableCleaning":
		pvcdef := newPvc(pv)
		pvc, err := r.k8sClient.Core().PersistentVolumeClaims("kube-system").Create(pvcdef)
		if err != nil {
			return microerror.Maskf(err, "failed to create persistent volume claim", pvc.Name)
		}
	case "BoundCleaning":
		pvcName := fmt.Sprintf("pv-cleaner-claim-%s", pv.Name)
		pvc, err := r.k8sClient.Core().PersistentVolumeClaims("kube-system").Get(pvcName, metav1.GetOptions{})
		if err != nil {
			return microerror.Maskf(err, "failed to get persistent volume claim", pvcName)
		}

		cleanupJobDef := newCleanupJob(pvc)
		cleanupJob, err := r.k8sClient.Batch().Jobs("kube-system").Create(cleanupJobDef)
		if errors.IsAlreadyExists(err) {
			cleanupJob, err = r.k8sClient.Batch().Jobs("kube-system").Get(cleanupJobDef.Name, metav1.GetOptions{})
			if err != nil {
				return microerror.Maskf(err, "failed to get cleanup job", pvc.Name)
			}
		} else if err != nil {
			return microerror.Maskf(err, "failed to create cleanup job", pvc.Name)
		}

		if cleanupJob.Status.Succeeded != 1 {
			r.logger.LogCtx(ctx, "job", cleanupJob.Name, "waiting for job to complete cleanup of pv", pv.Name)
			return nil
		}

		if err := r.k8sClient.Batch().Jobs("kube-system").Delete(cleanupJob.Name, &metav1.DeleteOptions{}); err != nil {
			return microerror.Maskf(err, "failed to delete cleanup job", cleanupJob.Name)
		}

		if err := r.k8sClient.Core().PersistentVolumeClaims("kube-system").Delete(pvcName, &metav1.DeleteOptions{}); err != nil {
			return microerror.Maskf(err, "failed to delete claim for persistent volume", pv.Name)
		}
	case "ReleasedCleaning":
		err = r.k8sClient.Core().PersistentVolumes().Delete(pv.Name, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

// newUpdateChange checks wherether persistent volume should be reconciled
// on update event.
func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	updatedVolume, err := toPV(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if reflect.DeepEqual(currentState, desiredState) {
		r.logger.LogCtx(ctx, "persistentvolume", updatedVolume.Name, "volume reconciled to desired state", "true")
		return nil, nil
	}

	return currentState, nil
}
