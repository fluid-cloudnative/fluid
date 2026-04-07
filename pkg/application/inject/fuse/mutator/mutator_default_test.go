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
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	applicationspod "github.com/fluid-cloudnative/fluid/pkg/utils/applications/pod"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

// buildScheme returns a runtime.Scheme with the required types registered.
func buildScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, datav1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, appsv1.AddToScheme(scheme))
	return scheme
}

// buildDefaultMutatorArgs creates a MutatorBuildArgs from pod + objects registered in scheme.
func buildDefaultMutatorArgs(t *testing.T, scheme *runtime.Scheme, podToMutate *corev1.Pod, objs ...runtime.Object) MutatorBuildArgs {
	t.Helper()
	c := fake.NewFakeClientWithScheme(scheme, objs...)
	pod, err := applicationspod.NewApplication(podToMutate).GetPodSpecs()
	require.NoError(t, err)
	require.Len(t, pod, 1)

	specs, err := CollectFluidObjectSpecs(pod[0])
	require.NoError(t, err)

	return MutatorBuildArgs{
		Client: c,
		Log:    fake.NullLogger(),
		Options: common.FuseSidecarInjectOption{
			EnableCacheDir:             false,
			SkipSidecarPostStartInject: false,
		},
		Specs: specs,
	}
}

func TestDefaultMutator_SinglePVC_BasicMutation(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	mountPropBidir := corev1.MountPropagationBidirectional
	mountPropH2C := corev1.MountPropagationHostToContainer

	assert.Len(t, podSpecs.Containers, 2)
	assert.True(t, len(podSpecs.Containers[0].Name) > 0 && podSpecs.Containers[0].Name[:len(common.FuseContainerName)] == common.FuseContainerName,
		"expected container name to have prefix %s, got %s", common.FuseContainerName, podSpecs.Containers[0].Name)

	assert.Contains(t, podSpecs.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:             "thin-fuse-mount-0",
		MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
		MountPropagation: &mountPropBidir,
	})
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

	expectedDataVol := corev1.Volume{
		Name: "data-vol-0",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", datasetNamespace, datasetName),
			},
		},
	}
	expectedCheckMountVol := corev1.Volume{
		Name: "default-check-mount-0",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: fmt.Sprintf("%s-default-check-mount", pv.Spec.CSI.VolumeAttributes[common.VolumeAttrMountType]),
				},
				DefaultMode: ptr.To[int32](0755),
			},
		},
	}
	assert.Contains(t, podSpecs.Volumes, expectedDataVol)
	assert.Contains(t, podSpecs.Volumes, expectedCheckMountVol)
}

func TestDefaultMutator_SkipSidecarPostStartInject(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)
	args.Options.SkipSidecarPostStartInject = true

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	assert.Len(t, podSpecs.Containers, 2)

	// No check-mount volume mount should be present when SkipSidecarPostStartInject=true
	for _, vm := range podSpecs.Containers[0].VolumeMounts {
		assert.NotEqual(t, "default-check-mount-0", vm.Name, "expected no check-mount-0 volume mount")
	}
	for _, vol := range podSpecs.Volumes {
		assert.NotEqual(t, "default-check-mount-0", vol.Name, "expected no check-mount-0 volume")
	}
}

func TestDefaultMutator_CustomizedDaemonSetFields(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)

	// Add customized fields to the FUSE daemonset
	daemonSet.Spec.Template.Spec.Containers[0].Env = append(
		daemonSet.Spec.Template.Spec.Containers[0].Env,
		corev1.EnvVar{Name: "FOO", Value: "BAR"},
	)
	daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(
		daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts,
		corev1.VolumeMount{Name: "myvol", MountPath: "/tmp/myvol"},
	)
	daemonSet.Spec.Template.Spec.Volumes = append(
		daemonSet.Spec.Template.Spec.Volumes,
		corev1.Volume{Name: "myvol", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/tmp/myvol"}}},
	)
	daemonSet.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4000m"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		},
	}

	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	assert.Len(t, podSpecs.Containers, 2)

	assert.Contains(t, podSpecs.Containers[0].VolumeMounts, corev1.VolumeMount{Name: "myvol-0", MountPath: "/tmp/myvol"})
	assert.Contains(t, podSpecs.Volumes, corev1.Volume{Name: "myvol-0", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/tmp/myvol"}}})
	assert.Contains(t, podSpecs.Containers[0].Env, corev1.EnvVar{Name: "FOO", Value: "BAR"})

	cpu2 := resource.MustParse("2")
	mem4Gi := resource.MustParse("4Gi")
	cpu4 := resource.MustParse("4")
	mem8Gi := resource.MustParse("8Gi")
	assert.Equal(t, cpu2.Cmp(podSpecs.Containers[0].Resources.Requests[corev1.ResourceCPU]), 0)
	assert.Equal(t, mem4Gi.Cmp(podSpecs.Containers[0].Resources.Requests[corev1.ResourceMemory]), 0)
	assert.Equal(t, cpu4.Cmp(podSpecs.Containers[0].Resources.Limits[corev1.ResourceCPU]), 0)
	assert.Equal(t, mem8Gi.Cmp(podSpecs.Containers[0].Resources.Limits[corev1.ResourceMemory]), 0)
}

func TestDefaultMutator_EnableCacheDir(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)

	// Add cache-dir volume to daemonset
	daemonSet.Spec.Template.Spec.Volumes = append(
		daemonSet.Spec.Template.Spec.Volumes,
		corev1.Volume{Name: "cache-dir", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/tmp/cache-dir"}}},
	)
	daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(
		daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts,
		corev1.VolumeMount{Name: "cache-dir", MountPath: "/tmp/cache-dir"},
	)

	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)
	args.Options.EnableCacheDir = true

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	assert.Len(t, podSpecs.Containers, 2)

	// When EnableCacheDir is true, cache related volumes should be kept
	foundVolumeMount := false
	for _, vm := range podSpecs.Containers[0].VolumeMounts {
		if vm.Name == "cache-dir-0" || vm.Name == "cache-dir" {
			foundVolumeMount = true
			break
		}
	}
	assert.True(t, foundVolumeMount, "expected a cache-dir volume mount to be present")

	foundVolume := false
	for _, vol := range podSpecs.Volumes {
		if vol.Name == "cache-dir-0" || vol.Name == "cache-dir" {
			foundVolume = true
			break
		}
	}
	assert.True(t, foundVolume, "expected a cache-dir volume to be present")
}

func TestDefaultMutator_FluidSubPath(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath] = "path-a"

	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

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

func TestDefaultMutator_InitContainerMountsPVC(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
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

	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	mountPropBidir := corev1.MountPropagationBidirectional
	mountPropH2C := corev1.MountPropagationHostToContainer

	// Should have 2 containers and 2 init containers
	assert.Len(t, podSpecs.Containers, 2)
	assert.Len(t, podSpecs.InitContainers, 2)

	assert.True(t, len(podSpecs.Containers[0].Name) >= len(common.FuseContainerName) &&
		podSpecs.Containers[0].Name[:len(common.FuseContainerName)] == common.FuseContainerName)
	assert.True(t, len(podSpecs.InitContainers[0].Name) >= len(common.InitFuseContainerName) &&
		podSpecs.InitContainers[0].Name[:len(common.InitFuseContainerName)] == common.InitFuseContainerName)

	// App sidecar: bidirectional mount
	assert.Contains(t, podSpecs.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:             "thin-fuse-mount-0",
		MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
		MountPropagation: &mountPropBidir,
	})

	// Init sidecar: bidirectional mount
	assert.Contains(t, podSpecs.InitContainers[0].VolumeMounts, corev1.VolumeMount{
		Name:             "thin-fuse-mount-0",
		MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
		MountPropagation: &mountPropBidir,
	})

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

func TestDefaultMutator_RandomSuffixMode_WithPodName(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	podToMutate.ObjectMeta.Annotations = map[string]string{
		common.HostMountPathModeOnDefaultPlatformKey: string(common.HostPathModeRandomSuffix),
	}

	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	assert.Len(t, podSpecs.Containers, 2)

	var fuseMountVol *corev1.Volume
	var dataVol *corev1.Volume
	for i := range podSpecs.Volumes {
		v := &podSpecs.Volumes[i]
		if v.Name == "thin-fuse-mount-0" {
			fuseMountVol = v
		}
		if v.Name == "data-vol-0" {
			dataVol = v
		}
	}

	require.NotNil(t, fuseMountVol)
	require.NotNil(t, dataVol)

	assert.Regexp(t, `^/runtime-mnt/thin/fluid/test-dataset//test-pod/\d+-[0-9a-z]{1,8}$`, fuseMountVol.HostPath.Path)
	assert.Regexp(t, `^/runtime-mnt/thin/fluid/test-dataset/test-pod/\d+-[0-9a-z]{1,8}/thin-fuse$`, dataVol.HostPath.Path)
}

func TestDefaultMutator_RandomSuffixMode_WithGenerateName(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	podToMutate.ObjectMeta.Annotations = map[string]string{
		common.HostMountPathModeOnDefaultPlatformKey: string(common.HostPathModeRandomSuffix),
	}
	podToMutate.ObjectMeta.GenerateName = "mypod-"
	podToMutate.ObjectMeta.Name = ""

	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	assert.Len(t, podSpecs.Containers, 2)

	var fuseMountVol *corev1.Volume
	var dataVol *corev1.Volume
	for i := range podSpecs.Volumes {
		v := &podSpecs.Volumes[i]
		if v.Name == "thin-fuse-mount-0" {
			fuseMountVol = v
		}
		if v.Name == "data-vol-0" {
			dataVol = v
		}
	}

	require.NotNil(t, fuseMountVol)
	require.NotNil(t, dataVol)

	assert.Regexp(t, `^/runtime-mnt/thin/fluid/test-dataset//mypod---generate-name/\d+-[0-9a-z]{1,8}$`, fuseMountVol.HostPath.Path)
	assert.Regexp(t, `^/runtime-mnt/thin/fluid/test-dataset/mypod---generate-name/\d+-[0-9a-z]{1,8}/thin-fuse$`, dataVol.HostPath.Path)
}

func TestDefaultMutator_NativeSidecar_AppContainerOnly(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)
	args.Options.SidecarInjectionMode = common.SidecarInjectionMode_NativeSidecar

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	assert.Len(t, podSpecs.Containers, 1)
	assert.Len(t, podSpecs.InitContainers, 1)

	assert.True(t, len(podSpecs.InitContainers[0].Name) >= len(common.FuseContainerName) &&
		podSpecs.InitContainers[0].Name[:len(common.FuseContainerName)] == common.FuseContainerName)

	containerRestartPolicyAlways := corev1.ContainerRestartPolicyAlways
	assert.Equal(t, &containerRestartPolicyAlways, podSpecs.InitContainers[0].RestartPolicy)
}

func TestDefaultMutator_NativeSidecar_BothContainersAndInitContainers(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	podToMutate.Spec.InitContainers = []corev1.Container{
		{
			Name:  "init-container",
			Image: "init-image",
			VolumeMounts: []corev1.VolumeMount{
				{Name: "data-vol-0", MountPath: "/data0"},
			},
		},
	}

	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, runtime, daemonSet, pv)
	args.Options.SidecarInjectionMode = common.SidecarInjectionMode_NativeSidecar

	mutator := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetName, datasetNamespace)
	require.NoError(t, err)

	require.NoError(t, mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))
	require.NoError(t, mutator.PostMutate())

	podSpecs := mutator.GetMutatedPodSpecs()
	require.NotNil(t, podSpecs)

	// Should inject ONLY ONE native sidecar into init containers
	assert.Len(t, podSpecs.Containers, 1)
	assert.Len(t, podSpecs.InitContainers, 2) // one native sidecar + one app init container

	assert.True(t, len(podSpecs.InitContainers[0].Name) >= len(common.FuseContainerName) &&
		podSpecs.InitContainers[0].Name[:len(common.FuseContainerName)] == common.FuseContainerName)

	containerRestartPolicyAlways := corev1.ContainerRestartPolicyAlways
	assert.Equal(t, &containerRestartPolicyAlways, podSpecs.InitContainers[0].RestartPolicy)
}

func TestDefaultMutator_MultiplePVCs(t *testing.T) {
	const datasetNum = 3

	scheme := buildScheme(t)
	datasetNames := make([]string, datasetNum)
	datasetNamespaces := make([]string, datasetNum)
	pvs := make([]*corev1.PersistentVolume, datasetNum)
	var objs []runtime.Object

	for i := 0; i < datasetNum; i++ {
		name := fmt.Sprintf("test-dataset-%d", i)
		ns := "fluid"
		datasetNames[i] = name
		datasetNamespaces[i] = ns

		dataset, runtimeObj, daemonSet, pv := test_buildFluidResources(name, ns)
		pvs[i] = pv
		objs = append(objs, dataset, runtimeObj, daemonSet, pv)
	}

	podToMutate := test_buildPodToMutate(datasetNames)
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, objs...)

	mutator := NewDefaultMutator(args)

	mountPropBidir := corev1.MountPropagationBidirectional
	mountPropH2C := corev1.MountPropagationHostToContainer

	for i := 0; i < datasetNum; i++ {
		runtimeInfo, err := base.GetRuntimeInfo(args.Client, datasetNames[i], datasetNamespaces[i])
		require.NoError(t, err)

		require.NoError(t, mutator.MutateWithRuntimeInfo(datasetNames[i], runtimeInfo, fmt.Sprintf("-%d", i)))

		podSpecs := mutator.GetMutatedPodSpecs()
		require.NotNil(t, podSpecs)

		assert.Len(t, podSpecs.Containers, i+2)

		assert.Contains(t, podSpecs.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:             fmt.Sprintf("thin-fuse-mount-%d", i),
			MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespaces[i], datasetNames[i]),
			MountPropagation: &mountPropBidir,
		})
		assert.Contains(t, podSpecs.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      fmt.Sprintf("default-check-mount-%d", i),
			ReadOnly:  true,
			MountPath: "/check-mount.sh",
			SubPath:   "check-mount.sh",
		})
		assert.Contains(t, podSpecs.Containers[len(podSpecs.Containers)-1].VolumeMounts, corev1.VolumeMount{
			Name:             fmt.Sprintf("data-vol-%d", i),
			MountPath:        fmt.Sprintf("/data%d", i),
			MountPropagation: &mountPropH2C,
		})
		assert.Contains(t, podSpecs.Volumes, corev1.Volume{
			Name: fmt.Sprintf("data-vol-%d", i),
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", datasetNamespaces[i], datasetNames[i]),
				},
			},
		})
		assert.Contains(t, podSpecs.Volumes, corev1.Volume{
			Name: fmt.Sprintf("default-check-mount-%d", i),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-default-check-mount", pvs[i].Spec.CSI.VolumeAttributes[common.VolumeAttrMountType]),
					},
					DefaultMode: ptr.To[int32](0755),
				},
			},
		})
	}

	require.NoError(t, mutator.PostMutate())
}

func TestPrepareFuseContainerPostStartScript_MatchingSHA(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, thinRuntime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, thinRuntime, daemonSet, pv)
	c := args.Client

	// First mutate: creates the ConfigMap
	mut := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(c, datasetName, datasetNamespace)
	require.NoError(t, err)
	require.NoError(t, mut.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))

	// Re-build fresh args/mutator to simulate a second webhook call
	pod2, err := applicationspod.NewApplication(podToMutate).GetPodSpecs()
	require.NoError(t, err)
	specs2, err := CollectFluidObjectSpecs(pod2[0])
	require.NoError(t, err)
	args2 := MutatorBuildArgs{Client: c, Log: fake.NullLogger(), Options: args.Options, Specs: specs2}
	mut2 := NewDefaultMutator(args2)
	runtimeInfo2, err := base.GetRuntimeInfo(c, datasetName, datasetNamespace)
	require.NoError(t, err)

	// Second mutate: SHA256 matches, should not error
	assert.NoError(t, mut2.MutateWithRuntimeInfo(datasetName, runtimeInfo2, "-0"))
}

func TestPrepareFuseContainerPostStartScript_StaleSHA(t *testing.T) {
	const (
		datasetName      = "test-dataset"
		datasetNamespace = "fluid"
	)
	scheme := buildScheme(t)
	dataset, thinRuntime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)
	podToMutate := test_buildPodToMutate([]string{datasetName})
	args := buildDefaultMutatorArgs(t, scheme, podToMutate, dataset, thinRuntime, daemonSet, pv)
	c := args.Client

	// First mutate: creates the ConfigMap
	mut := NewDefaultMutator(args)
	runtimeInfo, err := base.GetRuntimeInfo(c, datasetName, datasetNamespace)
	require.NoError(t, err)
	require.NoError(t, mut.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0"))

	// Deliberately corrupt the SHA256 annotation to simulate a stale configmap
	cmList := &corev1.ConfigMapList{}
	require.NoError(t, c.List(context.TODO(), cmList))
	for i := range cmList.Items {
		cm := &cmList.Items[i]
		if cm.Annotations != nil {
			if _, ok := cm.Annotations[common.AnnotationCheckMountScriptSHA256]; ok {
				cm.Annotations[common.AnnotationCheckMountScriptSHA256] = "deliberately-stale-sha"
				require.NoError(t, c.Update(context.TODO(), cm))
			}
		}
	}

	// Second mutate: SHA256 mismatch → should trigger update
	pod2, err := applicationspod.NewApplication(podToMutate).GetPodSpecs()
	require.NoError(t, err)
	specs2, err := CollectFluidObjectSpecs(pod2[0])
	require.NoError(t, err)
	args2 := MutatorBuildArgs{Client: c, Log: fake.NullLogger(), Options: args.Options, Specs: specs2}
	mut2 := NewDefaultMutator(args2)
	runtimeInfo2, err := base.GetRuntimeInfo(c, datasetName, datasetNamespace)
	require.NoError(t, err)
	require.NoError(t, mut2.MutateWithRuntimeInfo(datasetName, runtimeInfo2, "-0"))

	// Verify the SHA256 annotation was refreshed
	updatedCmList := &corev1.ConfigMapList{}
	require.NoError(t, c.List(context.TODO(), updatedCmList))
	for _, cm := range updatedCmList.Items {
		if cm.Annotations != nil {
			if sha, ok := cm.Annotations[common.AnnotationCheckMountScriptSHA256]; ok {
				assert.NotEqual(t, "deliberately-stale-sha", sha)
			}
		}
	}
}
