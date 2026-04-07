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

package mutator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestUnprivilegedMutator_SinglePVC_BasicMutation(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := runtime.NewScheme()
	require.NoError(t, datav1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, appsv1.AddToScheme(scheme))

	dataset, runtimeObj, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtimeObj, daemonSet, pv)

	mutator := NewUnprivilegedMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	mountPropBidir := corev1.MountPropagationBidirectional
	mountPropH2C := corev1.MountPropagationHostToContainer

	assert.Len(t, podSpecs.Containers, 2)
	assert.True(t, len(podSpecs.Containers[0].Name) >= len(common.FuseContainerName) &&
		podSpecs.Containers[0].Name[:len(common.FuseContainerName)] == common.FuseContainerName)

	// Unprivileged: not privileged, no SYS_ADMIN capability
	require.NotNil(t, podSpecs.Containers[0].SecurityContext)
	assert.Equal(t, ptr.To(false), podSpecs.Containers[0].SecurityContext.Privileged)
	for _, cap := range podSpecs.Containers[0].SecurityContext.Capabilities.Add {
		assert.NotEqual(t, corev1.Capability("SYS_ADMIN"), cap)
	}

	// Unprivileged: no bidirectional mount propagation on fuse sidecar
	for _, vm := range podSpecs.Containers[0].VolumeMounts {
		assert.False(t, vm.MountPropagation != nil && *vm.MountPropagation == mountPropBidir,
			"expected no bidirectional mount propagation on unprivileged sidecar volume mounts")
	}

	assert.Contains(t, podSpecs.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      "default-check-mount-0",
		ReadOnly:  true,
		MountPath: "/check-mount.sh",
		SubPath:   "check-mount.sh",
	})

	assert.Contains(t, podSpecs.Containers[1].VolumeMounts, corev1.VolumeMount{
		Name:             "data-vol-0",
		MountPath:        "/data0",
		MountPropagation: &mountPropH2C,
	})

	assert.Contains(t, podSpecs.Volumes, corev1.Volume{
		Name: "data-vol-0",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", datasetNamespace, datasetName),
			},
		},
	})
	assert.Contains(t, podSpecs.Volumes, corev1.Volume{
		Name: "default-check-mount-0",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: fmt.Sprintf("%s-default-check-mount", pv.Spec.CSI.VolumeAttributes[common.VolumeAttrMountType]),
				},
				DefaultMode: ptr.To[int32](0755),
			},
		},
	})
}

func TestUnprivilegedMutator_FluidSubPath(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := runtime.NewScheme()
	require.NoError(t, datav1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, appsv1.AddToScheme(scheme))

	dataset, runtimeObj, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath] = "path-a"

	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtimeObj, daemonSet, pv)

	mutator := NewUnprivilegedMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	assert.Len(t, podSpecs.Containers, 2)

	require.NotNil(t, podSpecs.Containers[0].SecurityContext)
	assert.Equal(t, ptr.To(false), podSpecs.Containers[0].SecurityContext.Privileged)
	for _, cap := range podSpecs.Containers[0].SecurityContext.Capabilities.Add {
		assert.NotEqual(t, corev1.Capability("SYS_ADMIN"), cap)
	}

	mountPropH2C := corev1.MountPropagationHostToContainer
	assert.Contains(t, podSpecs.Containers[1].VolumeMounts, corev1.VolumeMount{
		Name:             "data-vol-0",
		MountPath:        "/data0",
		MountPropagation: &mountPropH2C,
	})

	expectedHostPath := fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse/path-a", datasetNamespace, datasetName)
	assert.Contains(t, podSpecs.Volumes, corev1.Volume{
		Name: "data-vol-0",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: expectedHostPath,
			},
		},
	})
}

func TestUnprivilegedMutator_InitContainerMountsPVC(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := runtime.NewScheme()
	require.NoError(t, datav1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, appsv1.AddToScheme(scheme))

	dataset, runtimeObj, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})

	// Add an init container that also mounts the Fluid PVC
	podToMutate.Spec.InitContainers = []corev1.Container{
		{
			Name:  "init-container",
			Image: "init-image",
			VolumeMounts: []corev1.VolumeMount{
				{Name: "data-vol-0", MountPath: "/data0"},
			},
		},
	}

	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtimeObj, daemonSet, pv)

	mutator := NewUnprivilegedMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	assert.Len(t, podSpecs.Containers, 2)
	assert.Len(t, podSpecs.InitContainers, 2)

	assert.True(t, len(podSpecs.Containers[0].Name) >= len(common.FuseContainerName) &&
		podSpecs.Containers[0].Name[:len(common.FuseContainerName)] == common.FuseContainerName)
	assert.True(t, len(podSpecs.InitContainers[0].Name) >= len(common.InitFuseContainerName) &&
		podSpecs.InitContainers[0].Name[:len(common.InitFuseContainerName)] == common.InitFuseContainerName)

	// Both app sidecar and init sidecar should be unprivileged
	require.NotNil(t, podSpecs.Containers[0].SecurityContext)
	assert.Equal(t, ptr.To(false), podSpecs.Containers[0].SecurityContext.Privileged)
	for _, cap := range podSpecs.Containers[0].SecurityContext.Capabilities.Add {
		assert.NotEqual(t, corev1.Capability("SYS_ADMIN"), cap)
	}

	require.NotNil(t, podSpecs.InitContainers[0].SecurityContext)
	assert.Equal(t, ptr.To(false), podSpecs.InitContainers[0].SecurityContext.Privileged)
	for _, cap := range podSpecs.InitContainers[0].SecurityContext.Capabilities.Add {
		assert.NotEqual(t, corev1.Capability("SYS_ADMIN"), cap)
	}

	mountPropBidir := corev1.MountPropagationBidirectional
	mountPropH2C := corev1.MountPropagationHostToContainer

	// Unprivileged: no bidirectional mount propagation on app sidecar
	for _, vm := range podSpecs.Containers[0].VolumeMounts {
		assert.False(t, vm.MountPropagation != nil && *vm.MountPropagation == mountPropBidir &&
			vm.Name == "thin-fuse-mount-0",
			"app sidecar should not have bidirectional mount for thin-fuse-mount-0")
	}

	// Unprivileged: no bidirectional mount propagation on init sidecar
	for _, vm := range podSpecs.InitContainers[0].VolumeMounts {
		assert.False(t, vm.MountPropagation != nil && *vm.MountPropagation == mountPropBidir &&
			vm.Name == "thin-fuse-mount-0",
			"init sidecar should not have bidirectional mount for thin-fuse-mount-0")
	}

	// Init sidecar should have no lifecycle
	assert.Nil(t, podSpecs.InitContainers[0].Lifecycle)

	// App container: host-to-container mount
	assert.Contains(t, podSpecs.Containers[1].VolumeMounts, corev1.VolumeMount{
		Name:             "data-vol-0",
		MountPath:        "/data0",
		MountPropagation: &mountPropH2C,
	})

	// Init app container: host-to-container mount
	assert.Contains(t, podSpecs.InitContainers[1].VolumeMounts, corev1.VolumeMount{
		Name:             "data-vol-0",
		MountPath:        "/data0",
		MountPropagation: &mountPropH2C,
	})
}
