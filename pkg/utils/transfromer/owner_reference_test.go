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

package transfromer

import (
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestGenerateOwnerReferenceFromCRD(t *testing.T) {
	var (
		name      string                = "test-dataset"
		namespace string                = "fluid"
		dataset   *datav1alpha1.Dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				UID:       "12345",
			},
		}
		expect *common.OwnerReference = &common.OwnerReference{
			Enabled:            true,
			Controller:         true,
			BlockOwnerDeletion: false,
			UID:                "12345",
			Kind:               "Dataset",
			APIVersion:         "data.fluid.io/v1alpha1",
			Name:               name,
		}
	)

	var testScheme *runtime.Scheme = runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(testScheme)
	testScheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
	testObjs := []runtime.Object{}

	testObjs = append(testObjs, dataset.DeepCopy())

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	obj := &datav1alpha1.Dataset{}

	err := fakeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, obj)

	if err != nil {
		t.Errorf("Failed due to %v", err)
	}

	result := GenerateOwnerReferenceFromObject(obj)
	if !reflect.DeepEqual(result, expect) {
		t.Errorf("The expect %v, the result: %v, they are not equal", expect, result)
	}
}
