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
	"context"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/mutator"
	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestFuseInjector(t *testing.T) (*Injector, client.Client) {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	return &Injector{
		client: fakeClient,
		log:    logr.Discard(),
	}, fakeClient
}

// ---- injectCheckMountReadyScript tests ----

func TestInjectCheckMountReadyScript_NoRuntimeInfos(t *testing.T) {
	injector, _ := newTestFuseInjector(t)
	podSpecs := &mutator.MutatingPodSpecs{
		MetaObj:    metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
		Volumes:    []corev1.Volume{},
		Containers: []corev1.Container{},
	}

	err := injector.injectCheckMountReadyScript(podSpecs, map[string]base.RuntimeInfoInterface{})
	assert.NoError(t, err)
	assert.Len(t, podSpecs.Volumes, 0)
}

func TestInjectCheckMountReadyScript_WithRuntimeInfos(t *testing.T) {
	injector, _ := newTestFuseInjector(t)
	podSpecs := &mutator.MutatingPodSpecs{
		MetaObj: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
		Volumes: []corev1.Volume{
			{
				Name: "data-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "test-pvc"},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name: "app-container",
				VolumeMounts: []corev1.VolumeMount{
					{Name: "data-volume", MountPath: "/data"},
				},
			},
		},
	}
	runtimeInfo, err := base.BuildRuntimeInfo("test-pvc", "default", common.AlluxioRuntime)
	require.NoError(t, err)
	runtimeInfos := map[string]base.RuntimeInfoInterface{"test-pvc": runtimeInfo}

	err = injector.injectCheckMountReadyScript(podSpecs, runtimeInfos)
	assert.NoError(t, err)
	assert.Len(t, podSpecs.Volumes, 2)
	require.Len(t, podSpecs.Containers, 1)
	assert.Len(t, podSpecs.Containers[0].VolumeMounts, 2)
}

func TestInjectCheckMountReadyScript_GenerateName(t *testing.T) {
	injector, _ := newTestFuseInjector(t)
	podSpecs := &mutator.MutatingPodSpecs{
		MetaObj:    metav1.ObjectMeta{GenerateName: "test-pod-", Namespace: "default"},
		Volumes:    []corev1.Volume{},
		Containers: []corev1.Container{},
	}

	err := injector.injectCheckMountReadyScript(podSpecs, map[string]base.RuntimeInfoInterface{})
	assert.NoError(t, err)
}

func TestInjectCheckMountReadyScript_WithPostStartLabel(t *testing.T) {
	injector, _ := newTestFuseInjector(t)
	podSpecs := &mutator.MutatingPodSpecs{
		MetaObj: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels:    map[string]string{"fluid.io/enable-injection": "true"},
		},
		Volumes: []corev1.Volume{
			{
				Name: "data-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "test-pvc"},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:         "app-container",
				VolumeMounts: []corev1.VolumeMount{{Name: "data-volume", MountPath: "/data"}},
			},
		},
	}

	err := injector.injectCheckMountReadyScript(podSpecs, map[string]base.RuntimeInfoInterface{})
	assert.NoError(t, err)
}

func TestInjectCheckMountReadyScript_ExistingPostStart(t *testing.T) {
	injector, _ := newTestFuseInjector(t)
	existingPostStart := &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "echo existing"}},
	}
	podSpecs := &mutator.MutatingPodSpecs{
		MetaObj: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels:    map[string]string{"fluid.io/enable-injection": "true"},
		},
		Volumes: []corev1.Volume{
			{
				Name: "data-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "test-pvc"},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:         "app-container",
				VolumeMounts: []corev1.VolumeMount{{Name: "data-volume", MountPath: "/data"}},
				Lifecycle:    &corev1.Lifecycle{PostStart: existingPostStart},
			},
		},
	}

	err := injector.injectCheckMountReadyScript(podSpecs, map[string]base.RuntimeInfoInterface{})
	assert.NoError(t, err)
	assert.Equal(t, existingPostStart, podSpecs.Containers[0].Lifecycle.PostStart)
}

func TestInjectCheckMountReadyScript_InitContainers(t *testing.T) {
	injector, _ := newTestFuseInjector(t)
	podSpecs := &mutator.MutatingPodSpecs{
		MetaObj: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
		Volumes: []corev1.Volume{
			{
				Name: "data-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "test-pvc"},
				},
			},
		},
		InitContainers: []corev1.Container{
			{
				Name:         "init-container",
				VolumeMounts: []corev1.VolumeMount{{Name: "data-volume", MountPath: "/data"}},
			},
		},
		Containers: []corev1.Container{},
	}

	err := injector.injectCheckMountReadyScript(podSpecs, map[string]base.RuntimeInfoInterface{})
	assert.NoError(t, err)
}

// ---- ensureScriptConfigMapExists tests ----

func TestEnsureScriptConfigMapExists_Create(t *testing.T) {
	injector, fakeClient := newTestFuseInjector(t)

	appScriptGen, err := injector.ensureScriptConfigMapExists("default")
	require.NoError(t, err)
	assert.NotNil(t, appScriptGen)

	cm := appScriptGen.BuildConfigmap()
	retrievedCM := &corev1.ConfigMap{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, retrievedCM)
	assert.NoError(t, err)
}

func TestEnsureScriptConfigMapExists_MatchingSHA256(t *testing.T) {
	injector, fakeClient := newTestFuseInjector(t)
	appScriptGen := poststart.NewScriptGeneratorForApp("default")
	cm := appScriptGen.BuildConfigmap()
	err := fakeClient.Create(context.TODO(), cm)
	require.NoError(t, err)

	_, err = injector.ensureScriptConfigMapExists("default")
	assert.NoError(t, err)

	retrievedCM := &corev1.ConfigMap{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, retrievedCM)
	require.NoError(t, err)
	assert.Contains(t, retrievedCM.Annotations, common.AnnotationCheckMountScriptSHA256)
	assert.Equal(t, appScriptGen.GetScriptSHA256(), retrievedCM.Annotations[common.AnnotationCheckMountScriptSHA256])
}

func TestEnsureScriptConfigMapExists_StaleSHA256(t *testing.T) {
	injector, fakeClient := newTestFuseInjector(t)
	appScriptGen := poststart.NewScriptGeneratorForApp("default")
	cm := appScriptGen.BuildConfigmap()
	cm.Annotations[common.AnnotationCheckMountScriptSHA256] = "stale-sha256"
	cm.Data["check-fluid-mount-ready.sh"] = "old script content"
	err := fakeClient.Create(context.TODO(), cm)
	require.NoError(t, err)

	_, err = injector.ensureScriptConfigMapExists("default")
	assert.NoError(t, err)

	retrievedCM := &corev1.ConfigMap{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, retrievedCM)
	require.NoError(t, err)
	assert.Equal(t, appScriptGen.GetScriptSHA256(), retrievedCM.Annotations[common.AnnotationCheckMountScriptSHA256])
	assert.NotEqual(t, "old script content", retrievedCM.Data["check-fluid-mount-ready.sh"])
}

func TestEnsureScriptConfigMapExists_MissingSHA256Annotation(t *testing.T) {
	injector, fakeClient := newTestFuseInjector(t)
	appScriptGen := poststart.NewScriptGeneratorForApp("default")
	cm := appScriptGen.BuildConfigmap()
	delete(cm.Annotations, common.AnnotationCheckMountScriptSHA256)
	err := fakeClient.Create(context.TODO(), cm)
	require.NoError(t, err)

	_, err = injector.ensureScriptConfigMapExists("default")
	assert.NoError(t, err)

	retrievedCM := &corev1.ConfigMap{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, retrievedCM)
	require.NoError(t, err)
	assert.Contains(t, retrievedCM.Annotations, common.AnnotationCheckMountScriptSHA256)
	assert.Equal(t, appScriptGen.GetScriptSHA256(), retrievedCM.Annotations[common.AnnotationCheckMountScriptSHA256])
}

// ---- collectDatasetVolumeMountInfo tests ----

func TestCollectDatasetVolumeMountInfo_NonPVCVolume(t *testing.T) {
	volMounts := []corev1.VolumeMount{{Name: "empty-volume", MountPath: "/empty"}}
	volumes := []corev1.Volume{
		{
			Name:         "empty-volume",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		},
	}

	result := collectDatasetVolumeMountInfo(volMounts, volumes, map[string]base.RuntimeInfoInterface{})
	assert.Len(t, result, 0)
}

func TestCollectDatasetVolumeMountInfo_PVCNotInRuntimeInfos(t *testing.T) {
	volMounts := []corev1.VolumeMount{{Name: "data-volume", MountPath: "/data"}}
	volumes := []corev1.Volume{
		{
			Name: "data-volume",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "unknown-pvc"},
			},
		},
	}

	result := collectDatasetVolumeMountInfo(volMounts, volumes, map[string]base.RuntimeInfoInterface{})
	assert.Len(t, result, 0)
}

// ---- assembleMountInfos tests ----

func TestAssembleMountInfos_MultipleEntries(t *testing.T) {
	path2RuntimeTypeMap := map[string]string{
		"/data1": "alluxio",
		"/data2": "juicefs",
	}

	mountPathStr, mountTypeStr := assembleMountInfos(path2RuntimeTypeMap)
	assert.Contains(t, mountPathStr, "/data1")
	assert.Contains(t, mountPathStr, "/data2")
	assert.Contains(t, mountPathStr, ":")
	assert.Contains(t, mountTypeStr, "alluxio")
	assert.Contains(t, mountTypeStr, "juicefs")
	assert.Contains(t, mountTypeStr, ":")
}

func TestAssembleMountInfos_EmptyMap(t *testing.T) {
	mountPathStr, mountTypeStr := assembleMountInfos(map[string]string{})
	assert.Equal(t, "", mountPathStr)
	assert.Equal(t, "", mountTypeStr)
}

func TestAssembleMountInfos_SingleEntry(t *testing.T) {
	mountPathStr, mountTypeStr := assembleMountInfos(map[string]string{"/data": "alluxio"})
	assert.Equal(t, "/data", mountPathStr)
	assert.Equal(t, "alluxio", mountTypeStr)
}

// MockRuntimeInfo is a mock implementation of base.RuntimeInfoInterface for testing
type MockRuntimeInfo struct {
	namespace   string
	runtimeType string
}

func (m *MockRuntimeInfo) GetNamespace() string {
	return m.namespace
}

func (m *MockRuntimeInfo) GetRuntimeType() string {
	return m.runtimeType
}

func (m *MockRuntimeInfo) GetName() string {
	return "mock-runtime"
}

func (m *MockRuntimeInfo) IsExclusive() bool {
	return false
}

func (m *MockRuntimeInfo) GetAnnotations() map[string]string {
	return map[string]string{}
}

func (m *MockRuntimeInfo) GetCommonLabelName() string {
	return "fluid.io/dataset"
}
