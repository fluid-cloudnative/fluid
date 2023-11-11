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

package volume

import (
	"context"
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func TestDeleteFusePersistentVolume(t *testing.T) {
	runtimeInfoHbase, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testPVInputs := []*v1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluid-hadoop",
			Annotations: map[string]string{
				"CreatedBy": "fluid",
			},
		},
		Spec: v1.PersistentVolumeSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluid-hbase",
		},
		Spec: v1.PersistentVolumeSpec{},
	}}

	testPVs := []runtime.Object{}
	for _, pvInput := range testPVInputs {
		testPVs = append(testPVs, pvInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)
	var testCase = []struct {
		runtimeInfo     base.RuntimeInfoInterface
		expectedDeleted bool
		pvName          string
	}{
		{
			runtimeInfo:     runtimeInfoHadoop,
			pvName:          "fluid-hadoop",
			expectedDeleted: true,
		},
		{
			pvName:          "fluid-hbase",
			runtimeInfo:     runtimeInfoHbase,
			expectedDeleted: false,
		},
	}
	for _, test := range testCase {
		key := types.NamespacedName{
			Name: test.pvName,
		}
		pv := &v1.PersistentVolume{}
		var log = ctrl.Log.WithName("delete")
		err := client.Get(context.TODO(), key, pv)
		if err != nil {
			t.Errorf("Expect no error, but got %v", err)
		}
		err = DeleteFusePersistentVolume(client, test.runtimeInfo, log)
		if err != nil {
			t.Errorf("failed to call DeleteFusePersistentVolume due to %v", err)
		}

		err = client.Get(context.TODO(), key, pv)
		if apierrs.IsNotFound(err) != test.expectedDeleted {
			t.Errorf("testcase %s Expect deleted %v, but got err %v", test.pvName, test.expectedDeleted, err)
		}

	}
}

func TestDeleteFusePersistentVolumeIfExists(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hbase",
			Annotations: map[string]string{
				"CreatedBy": "fluid",
			},
		},
		Spec: v1.PersistentVolumeSpec{},
	}}

	testPVs := []runtime.Object{}
	for _, pvInput := range testPVInputs {
		testPVs = append(testPVs, pvInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)
	var testCase = []struct {
		pvName         string
		expectedResult v1.PersistentVolume
	}{
		{
			pvName: "hadoop",
			expectedResult: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "hbase",
					Annotations: map[string]string{
						"CreatedBy": "fluid",
					},
				},
				Spec: v1.PersistentVolumeSpec{},
			},
		},
		{
			pvName:         "hbase",
			expectedResult: v1.PersistentVolume{},
		},
	}
	for _, test := range testCase {
		var log = ctrl.Log.WithName("delete")
		_ = deleteFusePersistentVolumeIfExists(client, test.pvName, log)

		key := types.NamespacedName{
			Name: test.pvName,
		}
		pv := &v1.PersistentVolume{}
		err := client.Get(context.TODO(), key, pv)
		if !apierrs.IsNotFound(err) {
			t.Errorf("testcase %s failed to delete due to %v", test.pvName, err)
		}

	}
}

func TestDeleteFusePersistentVolumeClaim(t *testing.T) {
	runtimeInfoHbase, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfoForceDelete, err := base.BuildRuntimeInfo("force-delete", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testPVCInputs := []*v1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "hbase",
				Namespace:   "fluid",
				Finalizers:  []string{"kubernetes.io/pvc-protection"},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: v1.PersistentVolumeClaimSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "force-delete",
				Namespace:         "fluid",
				Finalizers:        []string{"kubernetes.io/pvc-protection"},
				Annotations:       common.ExpectedFluidAnnotations,
				DeletionTimestamp: &metav1.Time{Time: time.Now().Add(-35 * time.Second)},
			},
		},
	}

	testPVCs := []runtime.Object{}
	for _, pvInput := range testPVCInputs {
		testPVCs = append(testPVCs, pvInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPVCs...)

	var testCase = []struct {
		runtimeInfo base.RuntimeInfoInterface
		isErr       bool
	}{
		{
			runtimeInfo: runtimeInfoHadoop,
			isErr:       false,
		},
		{
			runtimeInfo: runtimeInfoHbase,
			isErr:       true,
		},
		{
			runtimeInfo: runtimeInfoForceDelete,
			isErr:       false,
		},
	}
	for _, test := range testCase {
		var log = ctrl.Log.WithName("delete")
		if err := DeleteFusePersistentVolumeClaim(client, test.runtimeInfo, log); test.isErr != (err != nil) {
			if test.isErr {
				t.Errorf("Expected got error, but got nil")
			} else {
				t.Errorf("Expected no error, but got error: %v", err)
			}
		}
	}
}
