/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"context"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// A client-less runtime (Mooncake) still gets a PV and a PVC: the volume path is
// deliberately left untouched by #6176, so that app pods that do mount the dataset keep
// working while client-less ones ignore it. This pins that the PV/PVC are created even
// when the client component is switched off, so a later change to volume.go that keys
// the volume off the client would fail here.
func TestCreateVolumeWithClientDisabled(t *testing.T) {
	t.Setenv(testutil.FluidUnitTestEnv, "true")

	runtimeObj := newCacheRuntimeForConfigMapTest()
	runtimeObj.Spec.Master.Disabled = false
	runtimeObj.Spec.Worker.Disabled = false
	runtimeObj.Spec.Client.Disabled = true
	dataset := newDatasetForConfigMapTest()

	fakeClient := fake.NewFakeClientWithScheme(newCacheEngineTestScheme(t), runtimeObj, dataset)
	engine := &CacheEngine{
		Client:      fakeClient,
		Log:         fake.NullLogger(),
		name:        "demo",
		namespace:   "default",
		runtimeType: common.CacheRuntime,
	}

	if err := engine.CreateVolume(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	runtimeInfo, err := engine.getRuntimeInfo()
	if err != nil {
		t.Fatalf("failed to get runtime info: %v", err)
	}

	// The PV is cluster scoped, but CreatePersistentVolumeForRuntime stamps a namespace on
	// it, and the fake client keys the object by that namespace. Hence the namespace here.
	pv := &corev1.PersistentVolume{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: runtimeInfo.GetPersistentVolumeName(), Namespace: "default"}, pv); err != nil {
		t.Fatalf("expected the persistent volume to be created, got %v", err)
	}
	// The PV points the CSI driver at the FUSE mount point, which is what a client-less
	// runtime does not use. It is still filled in, and still names this runtime.
	if got := pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidPath]; got != engine.getFuseMountPoint() {
		t.Fatalf("expected fluid path %q, got %q", engine.getFuseMountPoint(), got)
	}

	pvc := &corev1.PersistentVolumeClaim{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "demo", Namespace: "default"}, pvc); err != nil {
		t.Fatalf("expected the persistent volume claim to be created, got %v", err)
	}

	// And the volume is removable again on the same path.
	if err := engine.DeleteVolume(context.Background()); err != nil {
		t.Fatalf("expected no error deleting the volume, got %v", err)
	}
}
