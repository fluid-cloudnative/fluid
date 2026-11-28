/*
Copyright 2023 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestRemoveNotFoundDatasetRef(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)

	var virtualDataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-virtual",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					Name:       "hbase",
					MountPoint: "dataset://fluid/hbase",
				},
			},
		},
	}

	var datasetRefExist = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-ref-exist",
			Namespace: "fluid",
		},
		Status: datav1alpha1.DatasetStatus{
			DatasetRef: []string{
				"fluid/hbase-virtual",
			},
		},
	}

	var datasetRefNotExist = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-ref-not-exist",
			Namespace: "fluid",
		},
		Status: datav1alpha1.DatasetStatus{
			DatasetRef: []string{
				"fluid/hbase-virtual-not-exists",
			},
		},
	}

	testObjs := []runtime.Object{}
	testObjs = append(testObjs, &virtualDataset)
	fakeclient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	testcases := []struct {
		name     string
		dataset  datav1alpha1.Dataset
		expected int
	}{
		{
			name:     "reference-dataset-exist",
			dataset:  datasetRefExist,
			expected: 1,
		}, {
			name:     "reference-dataset-not-exist",
			dataset:  datasetRefNotExist,
			expected: 0,
		},
	}
	for _, testcase := range testcases {
		datasetRef, err := RemoveNotFoundDatasetRef(fakeclient, testcase.dataset, fake.NullLogger())
		if err != nil {
			t.Errorf("test %s expect no error, but get %v", testcase.name, err)
		}
		if len(datasetRef) != testcase.expected {
			t.Errorf("test %s expect %v datasetRef, but get %v", testcase.name, testcase.expected, len(datasetRef))
		}
	}
}

// TestRemoveNotFoundDatasetRefGetError verifies that when looking up a reference
// dataset fails with a non-NotFound error, RemoveNotFoundDatasetRef returns that
// error together with the original datasetRef list instead of panicking. Before
// the fix it dereferenced the nil dataset returned by GetDataset on the error
// path, crashing the dataset controller on any transient API error.
func TestRemoveNotFoundDatasetRefGetError(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)

	refList := []string{"fluid/hbase-virtual"}
	dataset := datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-ref",
			Namespace: "fluid",
		},
		Status: datav1alpha1.DatasetStatus{
			DatasetRef: refList,
		},
	}

	fakeBase := ctrlfake.NewClientBuilder().WithScheme(testScheme).Build()
	// Force every Get to fail with a non-NotFound error, simulating a transient
	// API server failure while resolving the reference dataset.
	client := interceptor.NewClient(fakeBase, interceptor.Funcs{
		Get: func(ctx context.Context, c ctrlclient.WithWatch, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
			return apierrors.NewInternalError(fmt.Errorf("simulated API server error"))
		},
	})

	datasetRef, err := RemoveNotFoundDatasetRef(client, dataset, fake.NullLogger())
	if err == nil {
		t.Fatal("expect a non-NotFound error to be propagated, but got nil")
	}
	if !reflect.DeepEqual(datasetRef, refList) {
		t.Errorf("expect the original datasetRef %v to be returned unchanged, but got %v", refList, datasetRef)
	}
}
