/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"context"
	"strconv"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

// TestJuiceFSEngine_CreateVolume is a unit test function that tests the CreateVolume method of JuiceFSEngine.
// The main purpose of this test function is to verify whether JuiceFSEngine can correctly create a PersistentVolume (PV) and PersistentVolumeClaim (PVC).
// The test process includes:
// 1. Building a RuntimeInfo object.
// 2. Creating a test Dataset object.
// 3. Initializing JuiceFSEngine with a fake client.
// 4. Calling the CreateVolume method to create PV and PVC.
// 5. Verifying whether PV and PVC are successfully created.
// If any step fails, the test will return an error.
//
// Parameters:
//   - t *testing.T: A testing.T object used to manage test state and support test log output.
//
// Returns:
//   - No return value. If the test fails, errors will be reported via t.Errorf.
func TestJuiceFSEngine_CreateVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", common.JuiceFSRuntime)
	runtimeInfo.SetOwnerDatasetUID("dummy-dataset-uid")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := &JuiceFSEngine{
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

// TestJuiceFSEngine_CreateFusePersistentVolume tests the createFusePersistentVolume function of the JuiceFSEngine.
// It verifies that the function correctly creates a PersistentVolume and checks the result.

func TestJuiceFSEngine_createFusePersistentVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", common.JuiceFSRuntime)
	runtimeInfo.SetOwnerDatasetUID("dummy-dataset-uid")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

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
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", common.JuiceFSRuntime)
	runtimeInfo.SetOwnerDatasetUID("dummy-dataset-uid")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	engine := &JuiceFSEngine{
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "test",
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
