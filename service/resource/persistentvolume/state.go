package persistentvolume

import (
	"context"
)

// GetCurrentState returns current state of the recycled persistent volume
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {

	return nil, nil
}

// GetDesiredState returns desired state of the recycled persistent volume
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, nil
}
