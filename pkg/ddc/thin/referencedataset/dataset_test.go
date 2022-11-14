package referencedataset

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetMountedDatasetNamespacedName(t *testing.T) {
	tests := []struct {
		virtualDataset *datav1alpha1.Dataset
		want           int
	}{
		{
			virtualDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://ns-a/n-a",
						},
						{
							MountPoint: "dataset://ns-b/n-b",
						},
					},
				},
			},
			want: 2,
		},
		{
			virtualDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://ns-a/n-a",
						},
						{
							MountPoint: "http://ns-b/n-b",
						},
					},
				},
			},
			want: 1,
		},
		{
			virtualDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		if got := getMountedDatasetNamespacedName(tt.virtualDataset); len(got) != tt.want {
			t.Errorf("getMountedDatasetNamespacedName() len = %v, want %v", got, tt.want)
		}
	}
}

func TestGetDatasetRefName(t *testing.T) {
	refNameA := getDatasetRefName("a-b", "c")
	refNameB := getDatasetRefName("a", "bc")

	if refNameB == refNameA {
		t.Errorf("RefName is equal for different name and namespace")
	}
}
