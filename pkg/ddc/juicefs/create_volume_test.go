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

package juicefs

import (
	"context"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

func TestJuiceFSEngine_CreateVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfo.SetupFuseDeployMode(false, nil)

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JuiceFSEngine{
		Client:      client,
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "hbase",
		runtimeInfo: runtimeInfo,
		runtime: &datav1alpha1.JuiceFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
		},
	}

	err = engine.CreateVolume()
	if err != nil {
		t.Errorf("fail to exec CreateVolume with error %v", err)
	}

	var pvs v1.PersistentVolumeList
	err = client.List(context.TODO(), &pvs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvs.Items) != 1 {
		t.Errorf("fail to create the pv")
	}

	var pvcs v1.PersistentVolumeClaimList
	err = client.List(context.TODO(), &pvcs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvcs.Items) != 1 {
		t.Errorf("fail to create the pvc")
	}
}

func TestJuiceFSEngine_createFusePersistentVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfo.SetupFuseDeployMode(false, nil)

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JuiceFSEngine{
		Client:      client,
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "test",
		runtimeInfo: runtimeInfo,
	}

	err = engine.createFusePersistentVolume()
	if err != nil {
		t.Errorf("fail to exec createFusePersistentVolume with error %v", err)
	}

	var pvs v1.PersistentVolumeList
	err = client.List(context.TODO(), &pvs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvs.Items) != 1 {
		t.Errorf("fail to create the pv")
	}
}

func TestJuiceFSEngine_createFusePersistentVolumeClaim(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfo.SetupFuseDeployMode(false, nil)

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JuiceFSEngine{
		Client:      client,
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "test",
		runtimeInfo: runtimeInfo,
	}

	err = engine.createFusePersistentVolumeClaim()
	if err != nil {
		t.Errorf("fail to exec createFusePersistentVolumeClaim with error %v", err)
	}

	var pvcs v1.PersistentVolumeClaimList
	err = client.List(context.TODO(), &pvcs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvcs.Items) != 1 {
		t.Errorf("fail to create the pvc")
	}
}
