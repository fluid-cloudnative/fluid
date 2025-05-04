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

package alluxio

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

// TestCreateVolume tests the CreateVolume function of the AlluxioEngine.
// It verifies that the function successfully creates a PersistentVolume (PV)
// and a PersistentVolumeClaim (PVC) for the given dataset. The test sets up
// a fake Kubernetes client with a mock dataset and checks if exactly one PV
// and one PVC are created after the function execution
func TestCreateVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := &AlluxioEngine{
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "hbase",
		runtimeInfo: runtimeInfo,
		runtime: &datav1alpha1.AlluxioRuntime{
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
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
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

	engine := &AlluxioEngine{
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

// TestCreateFusePersistentVolumeClaim tests the createFusePersistentVolumeClaim function of the AlluxioEngine.
// It ensures that the function successfully creates a PersistentVolumeClaim (PVC) with the correct metadata.
// The test sets up a fake Kubernetes client with a mock DaemonSet and a Dataset to simulate the environment.
// After invoking the target function, it verifies that a PVC is created and labeled with the correct
// fuse generation based on the input DaemonSet's generation.
func TestCreateFusePersistentVolumeClaim(t *testing.T) {
	// Prepare runtime information for the AlluxioEngine.
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// Initialize a fake AlluxioEngine instance with mock runtime info.
	engine := &AlluxioEngine{
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "hbase",
		runtimeInfo: runtimeInfo,
	}

	// Set the Fuse name based on runtime information.
	engine.runtimeInfo.SetFuseName(engine.getFuseName())

	// Create a mock DaemonSet representing the Fuse deployment.
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

	// Create a test Dataset to simulate the presence of an Alluxio dataset.
	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	// Build the fake client using the test scheme and input objects.
	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	testObjs = append(testObjs, testDsInputs)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine.Client = client

	// Call the method to create the PVC and assert no errors occurred.
	err = engine.createFusePersistentVolumeClaim()
	if err != nil {
		t.Errorf("fail to exec createFusePersistentVolumeClaim with error %v", err)
	}

	// List all PVCs in the fake client and verify that exactly one was created.
	var pvcs v1.PersistentVolumeClaimList
	err = client.List(context.TODO(), &pvcs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvcs.Items) != 1 {
		t.Errorf("fail to create the pvc")
	}

	// Validate that the PVC has the correct fuse generation label.
	if pvcs.Items[0].Labels[common.LabelRuntimeFuseGeneration] != strconv.Itoa(int(testDsInputs.Generation)) {
		t.Errorf("fail to check fuse generation on pvc")
	}
}
