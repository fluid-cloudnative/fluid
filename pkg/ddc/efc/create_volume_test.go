/*
  Copyright 2022 The Fluid Authors.

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

package efc

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestEFCEngine_CreateVolume(t *testing.T) {
	t.Setenv(utils.MountRoot, "/runtime-mnt")
	testRuntimeInputs := []*datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{},
		},
	}

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testDsInputs := []*appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-fuse",
				Namespace: "fluid",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "fuse-image:v1",
							},
						},
					},
				},
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, runtimeInput := range testRuntimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	for _, dsInput := range testDsInputs {
		testObjs = append(testObjs, dsInput)
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &EFCEngine{
		Client:    client,
		Log:       fake.NullLogger(),
		namespace: "fluid",
		name:      "hbase",
	}
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", common.EFCRuntime)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	engine.runtimeInfo = runtimeInfo
	engine.runtimeInfo.SetFuseName(engine.getFuseName())
	if err := engine.CreateVolume(context.Background()); err != nil {
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

func TestEFCEngine_createFusePersistentVolume(t *testing.T) {
	t.Setenv(utils.MountRoot, "/runtime-mnt")
	testRuntimeInputs := []*datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{},
		},
	}

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testObjs := []runtime.Object{}
	for _, runtimeInput := range testRuntimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &EFCEngine{
		Client:    client,
		Log:       fake.NullLogger(),
		namespace: "fluid",
		name:      "hbase",
	}

	err := engine.createFusePersistentVolume(context.Background())
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

func TestEFCEngine_createFusePersistentVolumeWithNodeSelector(t *testing.T) {
	t.Setenv(utils.MountRoot, "/runtime-mnt")
	runtimeInput := &datav1alpha1.EFCRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	datasetInput := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}

	client := fake.NewFakeClientWithScheme(testScheme, runtimeInput, datasetInput)
	engine := &EFCEngine{
		Client:    client,
		Log:       fake.NullLogger(),
		namespace: "fluid",
		name:      "hbase",
	}

	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", common.EFCRuntime)
	if err != nil {
		t.Fatalf("fail to create runtimeInfo: %v", err)
	}
	runtimeInfo.SetFuseNodeSelector(map[string]string{
		"zone":   "us-east-1a",
		"arch":   "amd64",
		"region": "us-east-1",
	})
	engine.runtimeInfo = runtimeInfo

	err = engine.createFusePersistentVolume(context.Background())
	if err != nil {
		t.Fatalf("fail to exec createFusePersistentVolume: %v", err)
	}

	var pvs v1.PersistentVolumeList
	if err = client.List(context.TODO(), &pvs); err != nil {
		t.Fatalf("fail to list pvs: %v", err)
	}
	if len(pvs.Items) != 1 {
		t.Fatalf("expected 1 pv, got %d", len(pvs.Items))
	}
	pv := pvs.Items[0]
	if pv.Spec.NodeAffinity == nil || pv.Spec.NodeAffinity.Required == nil || len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms) != 1 {
		t.Fatalf("expected 1 node selector term, got %v", pv.Spec.NodeAffinity)
	}
	expectedExpressions := []v1.NodeSelectorRequirement{
		{
			Key:      "arch",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"amd64"},
		},
		{
			Key:      "region",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"us-east-1"},
		},
		{
			Key:      "zone",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"us-east-1a"},
		},
	}
	if !reflect.DeepEqual(pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions, expectedExpressions) {
		t.Errorf("expected MatchExpressions %v, got %v", expectedExpressions, pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions)
	}
}

func TestEFCEngine_createFusePersistentVolumeClaim(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", common.ThinRuntime)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	testRuntimeInputs := []*datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{},
		},
	}

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testDsInputs := []*appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-fuse",
				Namespace: "fluid",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "fuse-image:v1",
							},
						},
					},
				},
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, runtimeInput := range testRuntimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	for _, dsInput := range testDsInputs {
		testObjs = append(testObjs, dsInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &EFCEngine{
		Client:      client,
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "hbase",
		runtimeInfo: runtimeInfo,
	}
	engine.runtimeInfo.SetFuseName(engine.getFuseName())

	if err := engine.createFusePersistentVolumeClaim(context.Background()); err != nil {
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

	if pvcs.Items[0].Labels[common.LabelRuntimeFuseGeneration] != strconv.Itoa(int(testDsInputs[0].Generation)) {
		t.Errorf("fail to check fuse generation on pvc")
	}
}
