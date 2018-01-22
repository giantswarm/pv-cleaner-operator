package persistentvolume

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_RecyclePersistentVolume_GetCurrentState(t *testing.T) {
	testCases := []struct {
		description                     string
		obj                             interface{}
		expectedRecyclePersistentVolume interface{}
	}{
		{
			description: "recycle annotation is empty, expected recycle volume with 'Recycled' annotation",
			obj: &apiv1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "TestPersistentVolume",
					Annotations: map[string]string{},
				},
				Status: apiv1.PersistentVolumeStatus{
					Phase: "Available",
				},
			},
			expectedRecyclePersistentVolume: &RecyclePersistentVolume{
				Name:         "TestPersistentVolume",
				State:        "Available",
				RecycleState: recycled,
			},
		},
		{
			description: "recycle annotation is 'Cleaning', expected recycle volume with `Cleaning` annotation",
			obj: &apiv1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "TestPersistentVolume",
					Annotations: map[string]string{
						"pv-cleaner-operator.giantswarm.io/volume-recycle-state": cleaning,
					},
				},
				Status: apiv1.PersistentVolumeStatus{
					Phase: "Bound",
				},
			},
			expectedRecyclePersistentVolume: &RecyclePersistentVolume{
				Name:         "TestPersistentVolume",
				State:        "Bound",
				RecycleState: cleaning,
			},
		},
		{
			description: "recycle annotation is `Recycled`, expected recycle volume with 'Recycled' annotation",
			obj: &apiv1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "TestPersistentVolume",
					Annotations: map[string]string{
						"pv-cleaner-operator.giantswarm.io/volume-recycle-state": recycled,
					},
				},
				Status: apiv1.PersistentVolumeStatus{
					Phase: "Released",
				},
			},
			expectedRecyclePersistentVolume: &RecyclePersistentVolume{
				Name:         "TestPersistentVolume",
				State:        "Released",
				RecycleState: recycled,
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.GetCurrentState(context.TODO(), tc.obj)
		if err != nil {
			t.Fatalf("case %d unexpected error returned getting desired state: %s\n", i+1, err)
		}

		if !reflect.DeepEqual(tc.expectedRecyclePersistentVolume, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.expectedRecyclePersistentVolume, result)
		}
	}
}

func Test_Resource_RecyclePersistentVolume_GetDesiredState(t *testing.T) {
	testCases := []struct {
		description                     string
		obj                             interface{}
		expectedRecyclePersistentVolume interface{}
	}{
		{
			description: "recycle annotation is empty, expected recycle volume with 'Recycled' annotation",
			obj: &apiv1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "TestPersistentVolume",
					Annotations: map[string]string{},
				},
				Status: apiv1.PersistentVolumeStatus{
					Phase: "Available",
				},
			},
			expectedRecyclePersistentVolume: &RecyclePersistentVolume{
				Name:         "TestPersistentVolume",
				State:        "Available",
				RecycleState: recycled,
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.GetDesiredState(context.TODO(), tc.obj)
		if err != nil {
			t.Fatalf("case %d unexpected error returned getting desired state: %s\n", i+1, err)
		}

		if !reflect.DeepEqual(tc.expectedRecyclePersistentVolume, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.expectedRecyclePersistentVolume, result)
		}
	}
}
