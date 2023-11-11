/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/
package referencedataset

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetRuntime(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)

	testObjs := []runtime.Object{}

	var dataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
			},
		},
	}

	var runtime = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}

	var virtualRuntimeWithNoDataset = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-no-dataset",
			Namespace: "fluid",
		},
	}

	testObjs = append(testObjs, &dataset, &runtime, &virtualRuntimeWithNoDataset)
	fakeclient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	testcases := []struct {
		name        string
		e           ReferenceDatasetEngine
		expectedErr bool
	}{
		{
			name: "getRuntimeInfo success",
			e: ReferenceDatasetEngine{
				Id:        "test1",
				Client:    fakeclient,
				Log:       fake.NullLogger(),
				name:      "hbase",
				namespace: "fluid",
			},
			expectedErr: false,
		}, {
			name: "dataset-not-found",
			e: ReferenceDatasetEngine{
				Id:        "test2",
				Client:    fakeclient,
				Log:       fake.NullLogger(),
				name:      "hbase-no-dataset",
				namespace: "fluid",
			},
			expectedErr: false,
		}, {
			name: "runtime-not-found",
			e: ReferenceDatasetEngine{
				Id:        "test3",
				Client:    fakeclient,
				Log:       fake.NullLogger(),
				name:      "hbase-no-runtime",
				namespace: "fluid",
			},
			expectedErr: true,
		},
	}

	for _, testcase := range testcases {
		_, err := testcase.e.getRuntimeInfo()
		hasError := err != nil
		if testcase.expectedErr != hasError {
			t.Errorf("test %s expect error %t, get error %v", testcase.name, testcase.expectedErr, err)
		}
	}
}

func TestGetPhysicalRuntimeInfo(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)

	testObjs := []runtime.Object{}

	var dataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
			},
		},
	}

	var runtime = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Status: datav1alpha1.RuntimeStatus{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
			},
		},
	}

	var virtualRuntimeWithNoDataset = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-no-dataset",
			Namespace: "fluid",
		},
		Status: datav1alpha1.RuntimeStatus{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
			},
		},
	}

	var virtualRuntimeWithMultipleDataset = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-multiple-mounts",
			Namespace: "fluid",
		},
		Status: datav1alpha1.RuntimeStatus{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
				{
					MountPoint: "dataset://big-data2/done2",
				},
			},
		},
	}

	var physicalDataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "done",
			Namespace: "big-data",
		},
		Status: datav1alpha1.DatasetStatus{
			Runtimes: []datav1alpha1.Runtime{
				{
					Name:      "done",
					Namespace: "big-data",
					Type:      common.AlluxioRuntime,
				},
			},
		},
	}

	var physicalRuntime = datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "done",
			Namespace: "big-data",
		},
	}

	testObjs = append(testObjs, &dataset, &runtime, &virtualRuntimeWithNoDataset, &virtualRuntimeWithMultipleDataset)
	testObjs = append(testObjs, &physicalDataset, &physicalRuntime)
	fakeclient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	testcases := []struct {
		name        string
		e           ReferenceDatasetEngine
		expectedErr bool
		notFountErr bool
	}{
		{
			name: "Dataset-exists",
			e: ReferenceDatasetEngine{
				Id:        "test1",
				Client:    fakeclient,
				Log:       fake.NullLogger(),
				name:      "hbase",
				namespace: "fluid",
			},
			expectedErr: false,
			notFountErr: false,
		}, {
			name: "Dataset-not-exists",
			e: ReferenceDatasetEngine{
				Id:        "test2",
				Client:    fakeclient,
				Log:       fake.NullLogger(),
				name:      "hbase-no-dataset",
				namespace: "fluid",
			},
			expectedErr: false,
			notFountErr: false,
		}, {
			name: "Dataset-and-runtime-not-exists",
			e: ReferenceDatasetEngine{
				Id:        "test3",
				Client:    fakeclient,
				Log:       fake.NullLogger(),
				name:      "hbase-not-exists",
				namespace: "fluid",
			},
			expectedErr: true,
			notFountErr: true,
		}, {
			name: "Dataset-with-multiple-ref-mountpoints",
			e: ReferenceDatasetEngine{
				Id:        "test4",
				Client:    fakeclient,
				Log:       fake.NullLogger(),
				name:      "hbase-multiple-mounts",
				namespace: "fluid",
			},
			expectedErr: true,
			notFountErr: false,
		},
	}

	for _, testcase := range testcases {
		_, err := testcase.e.getPhysicalRuntimeInfo()
		hasError := err != nil
		if testcase.expectedErr != hasError {
			t.Errorf("test %s expect error %t, get error %v", testcase.name, testcase.expectedErr, err)
		}
		if hasError {
			notFoundError := utils.IgnoreNotFound(err) == nil
			if notFoundError != testcase.notFountErr {
				t.Errorf("test %s expect not found error %t, get error %v", testcase.name, testcase.expectedErr, err)
			}
		}
	}
}
