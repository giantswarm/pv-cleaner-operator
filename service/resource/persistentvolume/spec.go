package persistentvolume

import apiv1 "k8s.io/api/core/v1"

// RecyclePersistentVolume reflects PersistentVolume
// with additional RecycleState.
type RecyclePersistentVolume struct {
	Name         string
	State        apiv1.PersistentVolumePhase
	RecycleState string
}
