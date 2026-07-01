/*
Copyright 2024 The Fluid Authors.

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
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestSyncPVFuseLabelKey reproduces issue #6089 and verifies the fix:
// When ownerDatasetUID changes (e.g. Dataset deleted and recreated), the fuse daemonset
// nodeSelector key diverges from the PV's mount_pod_node_selector_key, causing mount failures.
// syncPVFuseLabelKey should detect and correct the stale PV attribute.
func TestSyncPVFuseLabelKey(t *testing.T) {
	// Long namespace forces UID-based label key (namespace+name would exceed 63 chars)
	longNamespace := "very-very-very-very-very-very-very-very-very-long-namespace"
	name := "jfsdemo-2"

	datasetUID_A := types.UID("c86886e0-778d-49cb-9942-0e717038071e") // original Dataset UID
	datasetUID_B := types.UID("5ec58061-ebba-4105-8d34-5e7ab4cddff9") // recreated Dataset UID

	// Compute the stale label key written to PV when Dataset had UID_A
	runtimeInfoA, _ := base.BuildRuntimeInfo(name, longNamespace, common.JuiceFSRuntime)
	runtimeInfoA.SetOwnerDatasetUID(datasetUID_A)
	staleLabelKey := runtimeInfoA.GetFuseLabelName()

	// After Dataset recreation, engine has UID_B; daemonset nodeSelector uses new key
	runtimeInfoB, _ := base.BuildRuntimeInfo(name, longNamespace, common.JuiceFSRuntime)
	runtimeInfoB.SetOwnerDatasetUID(datasetUID_B)
	expectedLabelKey := runtimeInfoB.GetFuseLabelName()

	if staleLabelKey == expectedLabelKey {
		t.Skip("label keys are identical with different UIDs — cannot test divergence")
	}
	t.Logf("Stale PV key  (UID_A): %s", staleLabelKey)
	t.Logf("Expected key  (UID_B): %s", expectedLabelKey)

	// Create a PV with the stale key, simulating the pre-existing PV from before Dataset recreation
	pvName := runtimeInfoB.GetPersistentVolumeName()
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        pvName,
			Annotations: common.GetExpectedFluidAnnotations(),
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					VolumeAttributes: map[string]string{
						common.VolumeAttrMountPodNodeSelectorKey: staleLabelKey,
					},
				},
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, pv)

	engine := &JuiceFSEngine{
		Log:         fake.NullLogger(),
		Client:      fakeClient,
		namespace:   longNamespace,
		name:        name,
		runtimeType: common.JuiceFSRuntime,
		runtimeInfo: runtimeInfoB,
		UnitTest:    true,
	}

	if err := engine.syncPVFuseLabelKey(); err != nil {
		t.Fatalf("syncPVFuseLabelKey returned error: %v", err)
	}

	updatedPV, err := kubeclient.GetPersistentVolume(fakeClient, pvName)
	if err != nil {
		t.Fatalf("failed to get PV after syncPVFuseLabelKey: %v", err)
	}

	updatedKey := updatedPV.Spec.CSI.VolumeAttributes[common.VolumeAttrMountPodNodeSelectorKey]
	if updatedKey != expectedLabelKey {
		t.Errorf("fix for issue #6089 failed: PV mount_pod_node_selector_key not corrected\n  want: %s\n  got:  %s",
			expectedLabelKey, updatedKey)
	}
	t.Logf("PV key after sync: %s", updatedKey)
}
