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
		Obj                             interface{}
		ExpectedRecyclePersistentVolume interface{}
	}{
		{
			Obj: &apiv1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "TestPersistentVolume",
					Annotations: map[string]string{},
				},
				Status: apiv1.PersistentVolumeStatus{
					Phase: "Available",
				},
			},
			ExpectedRecyclePersistentVolume: &RecyclePersistentVolume{
				Name:         "TestPersistentVolume",
				State:        "Available",
				RecycleState: recycled,
			},
		},
		{
			Obj: &apiv1.PersistentVolume{
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
			ExpectedRecyclePersistentVolume: &RecyclePersistentVolume{
				Name:         "TestPersistentVolume",
				State:        "Bound",
				RecycleState: cleaning,
			},
		},
		{
			Obj: &apiv1.PersistentVolume{
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
			ExpectedRecyclePersistentVolume: &RecyclePersistentVolume{
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
		result, err := newResource.GetCurrentState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatalf("case %d unexpected error returned getting desired state: %s\n", i+1, err)
		}

		if !reflect.DeepEqual(tc.ExpectedRecyclePersistentVolume, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedRecyclePersistentVolume, result)
		}
	}
}

func Test_Resource_RecyclePersistentVolume_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj                             interface{}
		ExpectedRecyclePersistentVolume interface{}
	}{
		{
			Obj: &apiv1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "TestPersistentVolume",
					Annotations: map[string]string{},
				},
				Status: apiv1.PersistentVolumeStatus{
					Phase: "Available",
				},
			},
			ExpectedRecyclePersistentVolume: &RecyclePersistentVolume{
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
		result, err := newResource.GetDesiredState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatalf("case %d unexpected error returned getting desired state: %s\n", i+1, err)
		}

		if !reflect.DeepEqual(tc.ExpectedRecyclePersistentVolume, result) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedRecyclePersistentVolume, result)
		}
	}
}
