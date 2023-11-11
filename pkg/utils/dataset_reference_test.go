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

package utils

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
