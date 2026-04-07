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

package fuse

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

// testCaseContext holds objects needed for injector test cases.
type testCaseContext struct {
	in             *corev1.Pod
	datasets       []*datav1alpha1.Dataset
	pvs            []*corev1.PersistentVolume
	pvcs           []*corev1.PersistentVolumeClaim
	fuseDaemonsets []*appsv1.DaemonSet
}

func mockTestCaseContext(datasetNames []string, namespace string) *testCaseContext {
	mockedDatasets := []*datav1alpha1.Dataset{}
	for _, name := range datasetNames {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		mockedDatasets = append(mockedDatasets, dataset)
	}

	mockedPVs := []*corev1.PersistentVolume{}
	for _, dataset := range mockedDatasets {
		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s", dataset.Namespace, dataset.Name),
			},
			Spec: corev1.PersistentVolumeSpec{
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						Driver: "fuse.csi.fluid.io",
						VolumeAttributes: map[string]string{
							common.VolumeAttrFluidPath: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", dataset.Namespace, dataset.Name),
							common.VolumeAttrMountType: common.ThinRuntime,
						},
					},
				},
			},
		}
		mockedPVs = append(mockedPVs, pv)
	}

	mockedPVCs := []*corev1.PersistentVolumeClaim{}
	for _, dataset := range mockedDatasets {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dataset.Name,
				Namespace: dataset.Namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				VolumeName: fmt.Sprintf("%s-%s", dataset.Namespace, dataset.Name),
			},
		}
		mockedPVCs = append(mockedPVCs, pvc)
	}

	hostPathCharDev := corev1.HostPathCharDev
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	mockedFuses := []*appsv1.DaemonSet{}
	for _, dataset := range mockedDatasets {
		fuseDs := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-fuse", dataset.Name),
				Namespace: dataset.Namespace,
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:    "thin-fuse",
								Args:    []string{"-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000"},
								Command: []string{"/entrypoint.sh"},
								Image:   "test-thin-fuse:v1.0.0",
								SecurityContext: &corev1.SecurityContext{
									Privileged: ptr.To(true),
								},
								VolumeMounts: []corev1.VolumeMount{
									{Name: "data", MountPath: "/mnt/disk1"},
									{Name: "fuse-device", MountPath: "/dev/fuse"},
									{
										Name:      "thin-fuse-mount",
										MountPath: fmt.Sprintf("/runtime-mnt/thin/%s/%s/", dataset.Namespace, dataset.Name),
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "data",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{Path: "/runtime_mnt/dataset-conflict"},
								},
							},
							{
								Name: "fuse-device",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/dev/fuse",
										Type: &hostPathCharDev,
									},
								},
							},
							{
								Name: "thin-fuse-mount",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s", dataset.Namespace, dataset.Name),
										Type: &hostPathDirectoryOrCreate,
									},
								},
							},
						},
					},
				},
			},
		}
		mockedFuses = append(mockedFuses, fuseDs)
	}

	inPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: namespace,
			Labels: map[string]string{
				common.InjectServerless: common.True,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "test-image", Name: "test-container"},
			},
		},
	}

	for _, dataset := range mockedDatasets {
		inPod.Spec.Volumes = append(inPod.Spec.Volumes, corev1.Volume{
			Name: dataset.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: dataset.Name,
				},
			},
		})
		inPod.Spec.Containers[0].VolumeMounts = append(inPod.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      dataset.Name,
			MountPath: "/data-" + dataset.Name,
		})
	}

	return &testCaseContext{
		in:             inPod,
		datasets:       mockedDatasets,
		pvs:            mockedPVs,
		pvcs:           mockedPVCs,
		fuseDaemonsets: mockedFuses,
	}
}

func buildInjectorFromTestCtx(t *testing.T, testCtx *testCaseContext) (*Injector, client.Client) {
	t.Helper()
	runtimeObjs := []runtime.Object{}
	for _, obj := range testCtx.datasets {
		runtimeObjs = append(runtimeObjs, obj)
	}
	for _, obj := range testCtx.pvs {
		runtimeObjs = append(runtimeObjs, obj)
	}
	for _, obj := range testCtx.pvcs {
		runtimeObjs = append(runtimeObjs, obj)
	}
	for _, obj := range testCtx.fuseDaemonsets {
		runtimeObjs = append(runtimeObjs, obj)
	}
	c := fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, runtimeObjs...)
	return NewInjector(c), c
}

func assertInjectionCorrect(t *testing.T, out *corev1.Pod, fuseDs *appsv1.DaemonSet, dataset *datav1alpha1.Dataset) {
	t.Helper()

	var containerIdx int = -1
	var containerName string
	for k, v := range out.Labels {
		if strings.HasPrefix(k, common.LabelContainerDatasetNameKeyPrefix) && v == dataset.Name {
			containerName = strings.TrimPrefix(k, common.LabelContainerDatasetNameKeyPrefix)
			containerIdx = slices.IndexFunc(out.Spec.Containers, func(c corev1.Container) bool {
				return c.Name == containerName
			})
			break
		}
	}
	assert.NotEmpty(t, containerName)
	assert.NotEqual(t, -1, containerIdx)

	assert.Equal(t, fuseDs.Spec.Template.Spec.Containers[0].Image, out.Spec.Containers[containerIdx].Image)
	assert.Equal(t, ptr.To(true), out.Spec.Containers[containerIdx].SecurityContext.Privileged)

	assert.Subset(t, out.Spec.Containers[containerIdx].Command, fuseDs.Spec.Template.Spec.Containers[0].Command)
	assert.Subset(t, out.Spec.Containers[containerIdx].Args, fuseDs.Spec.Template.Spec.Containers[0].Args)

	fuseContainerSuffix := strings.TrimPrefix(containerName, common.FuseContainerName)

	assert.Equal(t, &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{
			Command: []string{"bash", "-c", fmt.Sprintf("time /check-mount.sh /runtime-mnt/thin/%s/%s/ thin ", dataset.Namespace, dataset.Name)},
		},
	}, out.Spec.Containers[containerIdx].Lifecycle.PostStart)

	for _, vm := range fuseDs.Spec.Template.Spec.Containers[0].VolumeMounts {
		vm.Name = vm.Name + fuseContainerSuffix
		assert.Contains(t, out.Spec.Containers[containerIdx].VolumeMounts, vm)
	}
	assert.Contains(t, out.Spec.Containers[containerIdx].VolumeMounts, corev1.VolumeMount{
		Name:      "default-check-mount" + fuseContainerSuffix,
		ReadOnly:  true,
		MountPath: "/check-mount.sh",
		SubPath:   "check-mount.sh",
	})

	for _, v := range fuseDs.Spec.Template.Spec.Volumes {
		v.Name = v.Name + fuseContainerSuffix
		assert.Contains(t, out.Spec.Volumes, v)
	}
	assert.Contains(t, out.Spec.Volumes, corev1.Volume{
		Name: "default-check-mount" + fuseContainerSuffix,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "thin-default-check-mount"},
				DefaultMode:          ptr.To[int32](0755),
			},
		},
	})

	assert.Contains(t, out.Spec.Volumes, corev1.Volume{
		Name:         dataset.Name,
		VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", dataset.Namespace, dataset.Name)}},
	})

	hostToContainerMountPropagation := corev1.MountPropagationHostToContainer
	assert.Contains(t, out.Spec.Containers[len(out.Spec.Containers)-1].VolumeMounts, corev1.VolumeMount{
		Name:             dataset.Name,
		MountPath:        "/data-" + dataset.Name,
		MountPropagation: &hostToContainerMountPropagation,
	})
}

func TestInjectPod_SingleDataset(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-1"}, "fluid-test")
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)

	fuseDs := testCtx.fuseDaemonsets[0]
	require.NoError(t, err)
	assert.Equal(t, "fluid-test", out.ObjectMeta.Labels[common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0"])
	assert.Equal(t, "dataset-1", out.ObjectMeta.Labels[common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0"])
	assert.Len(t, out.Spec.Containers, 2)

	assertInjectionCorrect(t, out, fuseDs, testCtx.datasets[0])
}

func TestInjectPod_UserSpecifiedFields(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-1"}, "fluid-test")
	testCtx.fuseDaemonsets[0].Spec.Template.Spec.Volumes = append(
		testCtx.fuseDaemonsets[0].Spec.Template.Spec.Volumes,
		corev1.Volume{Name: "new-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
	)
	testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].VolumeMounts = append(
		testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].VolumeMounts,
		corev1.VolumeMount{Name: "new-vol", MountPath: "/new-vol"},
	)
	testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].Env = append(
		testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].Env,
		corev1.EnvVar{Name: "new-env", Value: "new-env-value"},
	)
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)

	fuseDs := testCtx.fuseDaemonsets[0]
	require.NoError(t, err)
	assert.Equal(t, "fluid-test", out.ObjectMeta.Labels[common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0"])
	assert.Equal(t, "dataset-1", out.ObjectMeta.Labels[common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0"])
	assert.Len(t, out.Spec.Containers, 2)

	assertInjectionCorrect(t, out, fuseDs, testCtx.datasets[0])
}

func TestInjectPod_AlreadyInjected(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-1"}, "fluid-test")
	testCtx.in.ObjectMeta.Labels[common.InjectSidecarDone] = common.True
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)
	require.NoError(t, err)
	assert.Equal(t, testCtx.in, out)
}

func TestInjectPod_DuplicatePVC(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-1"}, "fluid-test")
	testCtx.in.Spec.Volumes = append(testCtx.in.Spec.Volumes, corev1.Volume{
		Name: "duplicate-pvc",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: testCtx.datasets[0].Name,
			},
		},
	})
	testCtx.in.Spec.Containers[0].VolumeMounts = append(testCtx.in.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      "duplicate-pvc",
		MountPath: "/duplicate-pvc",
	})
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)
	require.NoError(t, err)
	assert.Equal(t, "fluid-test", out.ObjectMeta.Labels[common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0"])
	assert.Equal(t, "dataset-1", out.ObjectMeta.Labels[common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0"])
	assert.Len(t, out.Spec.Containers, 2)
	assertInjectionCorrect(t, out, testCtx.fuseDaemonsets[0], testCtx.datasets[0])

	assert.Contains(t, out.Spec.Volumes, corev1.Volume{
		Name: "duplicate-pvc",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", testCtx.datasets[0].Namespace, testCtx.datasets[0].Name),
			},
		},
	})

	hostToContainerMountPropagation := corev1.MountPropagationHostToContainer
	assert.Contains(t, out.Spec.Containers[1].VolumeMounts, corev1.VolumeMount{
		Name:             "duplicate-pvc",
		MountPath:        "/duplicate-pvc",
		MountPropagation: &hostToContainerMountPropagation,
	})
}

func TestInjectPod_UnprivilegedMutator(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-1"}, "fluid-test")
	testCtx.in.Labels[common.InjectUnprivilegedFuseSidecar] = common.True
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)
	require.NoError(t, err)

	assert.Equal(t, "fluid-test", out.ObjectMeta.Labels[common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0"])
	assert.Equal(t, "dataset-1", out.ObjectMeta.Labels[common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0"])
	assert.Len(t, out.Spec.Containers, 2)

	fuseContainer := out.Spec.Containers[0]
	assert.True(t, strings.HasPrefix(fuseContainer.Name, common.FuseContainerName))
	assert.NotNil(t, fuseContainer.SecurityContext)
	if fuseContainer.SecurityContext.Privileged != nil {
		assert.False(t, *fuseContainer.SecurityContext.Privileged)
	}
	if fuseContainer.SecurityContext.Capabilities != nil {
		assert.NotContains(t, fuseContainer.SecurityContext.Capabilities.Add, corev1.Capability("SYS_ADMIN"))
	}
}

func TestInjectPod_InitContainerWithPVC(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-1"}, "fluid-test")
	testCtx.in.Spec.InitContainers = append(testCtx.in.Spec.InitContainers, corev1.Container{
		Name: "init-container",
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      testCtx.datasets[0].Name,
				MountPath: "/init-data-" + testCtx.datasets[0].Name,
			},
		},
	})
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)
	require.NoError(t, err)

	fuseDs := testCtx.fuseDaemonsets[0]
	assert.Len(t, out.Spec.InitContainers, 2)
	assert.True(t, strings.HasPrefix(out.Spec.InitContainers[0].Name, common.InitFuseContainerName))

	for _, vm := range fuseDs.Spec.Template.Spec.Containers[0].VolumeMounts {
		vm.Name = vm.Name + "-0"
		assert.Contains(t, out.Spec.Containers[0].VolumeMounts, vm)
	}

	assert.Contains(t, out.Spec.InitContainers[0].Command, "sleep")
	assert.Contains(t, out.Spec.InitContainers[0].Args, "2s")
}

func TestInjectPod_MultipleDatasets(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-1", "dataset-2", "dataset-3"}, "fluid-test")
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)

	require.NoError(t, err)
	assert.Equal(t, "fluid-test", out.ObjectMeta.Labels[common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0"])
	assert.Equal(t, "dataset-1", out.ObjectMeta.Labels[common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0"])
	assert.Equal(t, "fluid-test", out.ObjectMeta.Labels[common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-1"])
	assert.Equal(t, "dataset-2", out.ObjectMeta.Labels[common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-1"])
	assert.Equal(t, "fluid-test", out.ObjectMeta.Labels[common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-2"])
	assert.Equal(t, "dataset-3", out.ObjectMeta.Labels[common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-2"])
	assert.Len(t, out.Spec.Containers, 4)

	for i := 0; i < len(testCtx.datasets); i++ {
		assertInjectionCorrect(t, out, testCtx.fuseDaemonsets[i], testCtx.datasets[i])
	}
}

func TestInjectPod_RefDataset(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"root-dataset"}, "fluid-test")
	testCtx.datasets = append(testCtx.datasets, &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{Name: "sub-dataset", Namespace: "ref"},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://fluid-test/root-dataset/path-a"}},
		},
	})

	testCtx.pvcs = append(testCtx.pvcs, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "sub-dataset", Namespace: "ref"},
		Spec:       corev1.PersistentVolumeClaimSpec{VolumeName: "ref-sub-dataset"},
	})

	testCtx.pvs = append(testCtx.pvs, &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "ref-sub-dataset"},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					Driver: "fuse.csi.fluid.io",
					VolumeAttributes: map[string]string{
						common.VolumeAttrFluidPath:    "/runtime-mnt/thin/fluid-test/root-dataset/thin-fuse",
						common.VolumeAttrMountType:    common.ThinRuntime,
						common.VolumeAttrFluidSubPath: "path-a",
					},
				},
			},
		},
	})

	testCtx.in.Namespace = "ref"
	testCtx.in.Spec.Volumes = []corev1.Volume{
		{
			Name: "sub-dataset",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "sub-dataset"},
			},
		},
	}
	testCtx.in.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{Name: "sub-dataset", MountPath: "/data"},
	}

	subDatasetFuseDs := testCtx.fuseDaemonsets[0].DeepCopy()
	subDatasetFuseDs.Name = "sub-dataset-fuse"
	subDatasetFuseDs.Namespace = "ref"
	testCtx.fuseDaemonsets = append(testCtx.fuseDaemonsets, subDatasetFuseDs)

	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	info, err := base.BuildRuntimeInfo(testCtx.datasets[1].Name, testCtx.datasets[1].Namespace, common.ThinRuntime)
	require.NoError(t, err)
	info.SetAPIReader(c)
	runtimeInfos[testCtx.datasets[1].Name] = info

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)

	fuseDs := testCtx.fuseDaemonsets[1]
	containerIdx := 0
	refDataset := testCtx.datasets[1]
	rootDataset := testCtx.datasets[0]
	require.NoError(t, err)
	assert.Equal(t, "ref", out.ObjectMeta.Labels[common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0"])
	assert.Equal(t, "sub-dataset", out.ObjectMeta.Labels[common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0"])
	assert.Len(t, out.Spec.Containers, 2)

	assert.Subset(t, out.Spec.Containers[containerIdx].Command, fuseDs.Spec.Template.Spec.Containers[0].Command)
	assert.Subset(t, out.Spec.Containers[containerIdx].Args, fuseDs.Spec.Template.Spec.Containers[0].Args)

	for _, v := range fuseDs.Spec.Template.Spec.Volumes {
		v.Name = v.Name + "-0"
		assert.Contains(t, out.Spec.Volumes, v)
	}
	assert.Contains(t, out.Spec.Volumes, corev1.Volume{
		Name: "default-check-mount" + "-0",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "thin-default-check-mount"},
				DefaultMode:          ptr.To[int32](0755),
			},
		},
	})

	assert.Contains(t, out.Spec.Volumes, corev1.Volume{
		Name: refDataset.Name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse/%s", rootDataset.Namespace, rootDataset.Name, testCtx.pvs[1].Spec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath]),
			},
		},
	})
}

func TestInjectPod_FuseMetrics_ScrapeAll(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-with-fuse-metrics"}, "fluid-test")
	testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
		{Name: "thin-fuse-metrics", ContainerPort: 8080},
	}
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime, base.WithClientMetrics(datav1alpha1.ClientMetrics{
			Enabled:      true,
			ScrapeTarget: base.MountModeSelectAll,
		}))
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)
	require.NoError(t, err)

	assert.Equal(t, "true", out.Annotations[common.AnnotationPrometheusFuseMetricsScrapeKey])

	var fuseContainer *corev1.Container
	for i := range out.Spec.Containers {
		if out.Spec.Containers[i].Name == common.FuseContainerName+"-0" {
			fuseContainer = &out.Spec.Containers[i]
			break
		}
	}
	require.NotNil(t, fuseContainer)
	assert.Contains(t, fuseContainer.Ports, corev1.ContainerPort{Name: "thin-fuse-metrics", ContainerPort: 8080})
}

func TestInjectPod_FuseMetrics_MountPodOnly(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-with-mountpod-scrape-target"}, "fluid-test")
	testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
		{Name: "thin-fuse-metrics", ContainerPort: 8080},
	}
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime, base.WithClientMetrics(datav1alpha1.ClientMetrics{
			Enabled:      true,
			ScrapeTarget: string(base.MountPodMountMode),
		}))
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)
	require.NoError(t, err)

	assert.NotContains(t, out.Annotations, common.AnnotationPrometheusFuseMetricsScrapeKey)

	var fuseContainer *corev1.Container
	for i := range out.Spec.Containers {
		if out.Spec.Containers[i].Name == common.FuseContainerName+"-0" {
			fuseContainer = &out.Spec.Containers[i]
			break
		}
	}
	require.NotNil(t, fuseContainer)
	assert.NotContains(t, fuseContainer.Ports, corev1.ContainerPort{Name: "thin-fuse-metrics", ContainerPort: 8080})
}

func TestInjectPod_FuseMetrics_SidecarOnly(t *testing.T) {
	testCtx := mockTestCaseContext([]string{"dataset-with-sidecar-scrape-target"}, "fluid-test")
	testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
		{Name: "thin-fuse-metrics", ContainerPort: 8080},
	}
	injector, c := buildInjectorFromTestCtx(t, testCtx)

	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, dataset := range testCtx.datasets {
		info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime, base.WithClientMetrics(datav1alpha1.ClientMetrics{
			Enabled:      true,
			ScrapeTarget: string(base.SidecarMountMode),
		}))
		require.NoError(t, err)
		info.SetAPIReader(c)
		runtimeInfos[dataset.Name] = info
	}

	out, err := injector.InjectPod(testCtx.in, runtimeInfos)
	require.NoError(t, err)

	assert.Equal(t, "true", out.Annotations[common.AnnotationPrometheusFuseMetricsScrapeKey])

	var fuseContainer *corev1.Container
	for i := range out.Spec.Containers {
		if out.Spec.Containers[i].Name == common.FuseContainerName+"-0" {
			fuseContainer = &out.Spec.Containers[i]
			break
		}
	}
	require.NotNil(t, fuseContainer)
	assert.Contains(t, fuseContainer.Ports, corev1.ContainerPort{Name: "thin-fuse-metrics", ContainerPort: 8080})
}
