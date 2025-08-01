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

package jindocache

import (
	"context"
	"strconv"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCreateVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", common.JindoRuntime)
	runtimeInfo.SetOwnerDatasetUID("dummy-dataset-uid")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	engine := &JindoCacheEngine{
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "hbase",
		runtimeInfo: runtimeInfo,
		runtime: &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
	}

	engine.runtimeInfo.SetFuseName(engine.getFuseName())

	testDsInputs := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.getFuseName(),
			Namespace: engine.namespace,
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
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	testObjs = append(testObjs, testDsInputs)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine.Client = client

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

	if pvcs.Items[0].Labels[common.LabelRuntimeFuseGeneration] != strconv.Itoa(int(testDsInputs.Generation)) {
		t.Errorf("fail to check fuse generation on pvc")
	}
}

func TestCreateFusePersistentVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", common.JindoRuntime)
	runtimeInfo.SetOwnerDatasetUID("dummy-dataset-uid")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
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
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JindoCacheEngine{
		Client:      client,
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "hbase",
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

func TestCreateFusePersistentVolumeClaim(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", common.JindoRuntime)
	runtimeInfo.SetOwnerDatasetUID("dummy-dataset-uid")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := &JindoCacheEngine{
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "hbase",
		runtimeInfo: runtimeInfo,
	}

	engine.runtimeInfo.SetFuseName(engine.getFuseName())

	testDsInputs := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engine.getFuseName(),
			Namespace: engine.namespace,
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
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	testObjs = append(testObjs, testDsInputs)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine.Client = client

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

	if pvcs.Items[0].Labels[common.LabelRuntimeFuseGeneration] != strconv.Itoa(int(testDsInputs.Generation)) {
		t.Errorf("fail to check fuse generation on pvc")
	}
}
