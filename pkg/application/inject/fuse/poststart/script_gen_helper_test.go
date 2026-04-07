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

package poststart

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// --- scriptGeneratorHelper.BuildConfigMap ---

func TestScriptGeneratorHelper_BuildConfigMap_Basic(t *testing.T) {
	helper := &scriptGeneratorHelper{
		configMapName:   "test-config",
		scriptContent:   "#!/bin/bash\necho 'test'",
		scriptFileName:  "init.sh",
		scriptMountPath: "/scripts",
	}
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			UID:       "test-uid-123",
		},
	}
	configMapKey := types.NamespacedName{
		Name:      "my-configmap",
		Namespace: "default",
	}

	cm := helper.BuildConfigMap(dataset, configMapKey)

	assert.Equal(t, "my-configmap", cm.Name)
	assert.Equal(t, "default", cm.Namespace)
	assert.Equal(t, "#!/bin/bash\necho 'test'", cm.Data["init.sh"])
	assert.Equal(t, "default-test-dataset", cm.Labels[common.LabelAnnotationDatasetId])
	assert.Contains(t, cm.Annotations, common.AnnotationCheckMountScriptSHA256)
}

func TestScriptGeneratorHelper_BuildConfigMap_DifferentNamespace(t *testing.T) {
	helper := &scriptGeneratorHelper{
		configMapName:   "poststart-script",
		scriptContent:   "echo 'hello world'",
		scriptFileName:  "startup.sh",
		scriptMountPath: "/opt/scripts",
	}
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-dataset",
			Namespace: "prod-ns",
			UID:       "uid-456",
		},
	}
	configMapKey := types.NamespacedName{
		Name:      "prod-configmap",
		Namespace: "prod-ns",
	}

	cm := helper.BuildConfigMap(dataset, configMapKey)

	assert.Equal(t, "prod-configmap", cm.Name)
	assert.Equal(t, "prod-ns", cm.Namespace)
	assert.Equal(t, "echo 'hello world'", cm.Data["startup.sh"])
	assert.Equal(t, "prod-ns-my-dataset", cm.Labels[common.LabelAnnotationDatasetId])
}

func TestScriptGeneratorHelper_BuildConfigMap_EmptyScriptContent(t *testing.T) {
	helper := &scriptGeneratorHelper{
		configMapName:   "empty-script",
		scriptContent:   "",
		scriptFileName:  "empty.sh",
		scriptMountPath: "/scripts",
	}
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "empty-dataset",
			Namespace: "test",
			UID:       "empty-uid",
		},
	}
	configMapKey := types.NamespacedName{
		Name:      "empty-cm",
		Namespace: "test",
	}

	cm := helper.BuildConfigMap(dataset, configMapKey)

	assert.Equal(t, "empty-cm", cm.Name)
	assert.Equal(t, "test", cm.Namespace)
	assert.Equal(t, "", cm.Data["empty.sh"])
	assert.Equal(t, "test-empty-dataset", cm.Labels[common.LabelAnnotationDatasetId])
}

// --- scriptGeneratorHelper.GetNamespacedConfigMapKey ---

func TestScriptGeneratorHelper_GetNamespacedConfigMapKey_Alluxio(t *testing.T) {
	helper := &scriptGeneratorHelper{configMapName: "poststart-script"}
	datasetKey := types.NamespacedName{Name: "test-dataset", Namespace: "default"}

	key := helper.GetNamespacedConfigMapKey(datasetKey, "Alluxio")

	assert.Equal(t, "alluxio-poststart-script", key.Name)
	assert.Equal(t, "default", key.Namespace)
}

func TestScriptGeneratorHelper_GetNamespacedConfigMapKey_JuiceFS(t *testing.T) {
	helper := &scriptGeneratorHelper{configMapName: "init-config"}
	datasetKey := types.NamespacedName{Name: "my-dataset", Namespace: "prod"}

	key := helper.GetNamespacedConfigMapKey(datasetKey, "JuiceFS")

	assert.Equal(t, "juicefs-init-config", key.Name)
	assert.Equal(t, "prod", key.Namespace)
}

func TestScriptGeneratorHelper_GetNamespacedConfigMapKey_LowercaseRuntime(t *testing.T) {
	helper := &scriptGeneratorHelper{configMapName: "script"}
	datasetKey := types.NamespacedName{Name: "dataset", Namespace: "ns"}

	key := helper.GetNamespacedConfigMapKey(datasetKey, "jindo")

	assert.Equal(t, "jindo-script", key.Name)
	assert.Equal(t, "ns", key.Namespace)
}

func TestScriptGeneratorHelper_GetNamespacedConfigMapKey_MixedCaseRuntime(t *testing.T) {
	helper := &scriptGeneratorHelper{configMapName: "my-config"}
	datasetKey := types.NamespacedName{Name: "test", Namespace: "test-ns"}

	key := helper.GetNamespacedConfigMapKey(datasetKey, "GooseFS")

	assert.Equal(t, "goosefs-my-config", key.Name)
	assert.Equal(t, "test-ns", key.Namespace)
}

// --- scriptGeneratorHelper.GetVolume ---

func TestScriptGeneratorHelper_GetVolume_Basic(t *testing.T) {
	helper := &scriptGeneratorHelper{configMapName: "test-config"}
	configMapKey := types.NamespacedName{Name: "my-configmap", Namespace: "default"}

	volume := helper.GetVolume(configMapKey)

	assert.Equal(t, "test-config", volume.Name)
	require.NotNil(t, volume.VolumeSource.ConfigMap)
	assert.Equal(t, "my-configmap", volume.VolumeSource.ConfigMap.Name)
	require.NotNil(t, volume.VolumeSource.ConfigMap.DefaultMode)
	assert.Equal(t, int32(0755), *volume.VolumeSource.ConfigMap.DefaultMode)
}

func TestScriptGeneratorHelper_GetVolume_DifferentConfigMap(t *testing.T) {
	helper := &scriptGeneratorHelper{configMapName: "poststart-vol"}
	configMapKey := types.NamespacedName{Name: "prod-cm", Namespace: "production"}

	volume := helper.GetVolume(configMapKey)

	assert.Equal(t, "poststart-vol", volume.Name)
	require.NotNil(t, volume.VolumeSource.ConfigMap)
	assert.Equal(t, "prod-cm", volume.VolumeSource.ConfigMap.Name)
	require.NotNil(t, volume.VolumeSource.ConfigMap.DefaultMode)
	assert.Equal(t, int32(0755), *volume.VolumeSource.ConfigMap.DefaultMode)
}

// --- scriptGeneratorHelper.GetVolumeMount ---

func TestScriptGeneratorHelper_GetVolumeMount_Basic(t *testing.T) {
	helper := &scriptGeneratorHelper{
		configMapName:   "test-config",
		scriptFileName:  "init.sh",
		scriptMountPath: "/scripts/init.sh",
	}

	vm := helper.GetVolumeMount()

	assert.Equal(t, "test-config", vm.Name)
	assert.Equal(t, "/scripts/init.sh", vm.MountPath)
	assert.Equal(t, "init.sh", vm.SubPath)
	assert.True(t, vm.ReadOnly)
}

func TestScriptGeneratorHelper_GetVolumeMount_DifferentPath(t *testing.T) {
	helper := &scriptGeneratorHelper{
		configMapName:   "poststart-config",
		scriptFileName:  "startup.sh",
		scriptMountPath: "/opt/scripts/startup.sh",
	}

	vm := helper.GetVolumeMount()

	assert.Equal(t, "poststart-config", vm.Name)
	assert.Equal(t, "/opt/scripts/startup.sh", vm.MountPath)
	assert.Equal(t, "startup.sh", vm.SubPath)
	assert.True(t, vm.ReadOnly)
}

func TestScriptGeneratorHelper_GetVolumeMount_RootPath(t *testing.T) {
	helper := &scriptGeneratorHelper{
		configMapName:   "root-config",
		scriptFileName:  "run.sh",
		scriptMountPath: "/run.sh",
	}

	vm := helper.GetVolumeMount()

	assert.Equal(t, "root-config", vm.Name)
	assert.Equal(t, "/run.sh", vm.MountPath)
	assert.Equal(t, "run.sh", vm.SubPath)
	assert.True(t, vm.ReadOnly)
}

// --- ScriptGeneratorForApp ---

func TestScriptGeneratorForApp_New(t *testing.T) {
	g := NewScriptGeneratorForApp("test-ns")
	require.NotNil(t, g)
	assert.Equal(t, "test-ns", g.namespace)
}

func TestScriptGeneratorForApp_BuildConfigmap_NameAndNamespace(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	cm := g.BuildConfigmap()

	assert.Equal(t, appConfigMapName, cm.Name)
	assert.Equal(t, "default", cm.Namespace)
	assert.Contains(t, cm.Data, appScriptName)
	assert.NotEmpty(t, cm.Data[appScriptName])
}

func TestScriptGeneratorForApp_BuildConfigmap_SHA256Annotation(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	cm := g.BuildConfigmap()

	assert.Contains(t, cm.Annotations, common.AnnotationCheckMountScriptSHA256)
	assert.Equal(t, appScriptContentSHA256, cm.Annotations[common.AnnotationCheckMountScriptSHA256])
}

func TestScriptGeneratorForApp_BuildConfigmap_SHA256LengthConstraint(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	cm := g.BuildConfigmap()

	sha256Annotation := cm.Annotations[common.AnnotationCheckMountScriptSHA256]
	assert.LessOrEqual(t, len(sha256Annotation), 63)
}

func TestScriptGeneratorForApp_BuildConfigmap_ConsistentAcrossNamespaces(t *testing.T) {
	g1 := NewScriptGeneratorForApp("ns-a")
	g2 := NewScriptGeneratorForApp("ns-b")
	cm1 := g1.BuildConfigmap()
	cm2 := g2.BuildConfigmap()

	assert.Equal(t, cm1.Name, cm2.Name)
	assert.Equal(t, "ns-a", cm1.Namespace)
	assert.Equal(t, "ns-b", cm2.Namespace)
	assert.Equal(t, cm1.Data, cm2.Data)
	assert.Equal(t, cm1.Labels, cm2.Labels)
}

func TestScriptGeneratorForApp_GetScriptSHA256_MatchesPackageConst(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	assert.Equal(t, appScriptContentSHA256, g.GetScriptSHA256())
}

func TestScriptGeneratorForApp_GetScriptSHA256_LengthConstraint(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	sha := g.GetScriptSHA256()
	assert.NotEmpty(t, sha)
	assert.LessOrEqual(t, len(sha), 63)
}

func TestScriptGeneratorForApp_GetPostStartCommand_ExecCommand(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	handler := g.GetPostStartCommand("/data1:/data2", "alluxio:jindo")

	require.NotNil(t, handler)
	require.NotNil(t, handler.Exec)
	expectedCmd := fmt.Sprintf("time %s %s %s", appScriptPath, "/data1:/data2", "alluxio:jindo")
	assert.Equal(t, []string{"bash", "-c", expectedCmd}, handler.Exec.Command)
}

func TestScriptGeneratorForApp_GetPostStartCommand_ScriptPath(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	handler := g.GetPostStartCommand("/mnt/data", "juicefs")

	require.NotNil(t, handler)
	require.NotNil(t, handler.Exec)
	assert.Contains(t, handler.Exec.Command[2], appScriptPath)
}

func TestScriptGeneratorForApp_GetVolume_Name(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	vol := g.GetVolume()

	assert.Equal(t, appVolName, vol.Name)
	require.NotNil(t, vol.VolumeSource.ConfigMap)
	assert.Equal(t, appConfigMapName, vol.VolumeSource.ConfigMap.Name)
}

func TestScriptGeneratorForApp_GetVolume_DefaultMode(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	vol := g.GetVolume()

	require.NotNil(t, vol.VolumeSource.ConfigMap.DefaultMode)
	assert.Equal(t, int32(0755), *vol.VolumeSource.ConfigMap.DefaultMode)
}

func TestScriptGeneratorForApp_GetVolumeMount_Properties(t *testing.T) {
	g := NewScriptGeneratorForApp("default")
	vm := g.GetVolumeMount()

	assert.Equal(t, appVolName, vm.Name)
	assert.Equal(t, appScriptPath, vm.MountPath)
	assert.Equal(t, appScriptName, vm.SubPath)
	assert.True(t, vm.ReadOnly)
}

func TestScriptGeneratorForApp_AppScriptContentSHA256_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, appScriptContentSHA256)
}

func TestScriptGeneratorForApp_AppScriptContentSHA256_MatchesComputed(t *testing.T) {
	expected := computeScriptSHA256(replacer.Replace(contentCheckMountReadyScript))
	assert.Equal(t, expected, appScriptContentSHA256)
}

func TestScriptGeneratorForApp_AppScriptContentSHA256_Length63(t *testing.T) {
	assert.Equal(t, 63, len(appScriptContentSHA256))
}

func TestScriptGeneratorForApp_BuildConfigmap_SHA256ConsistentWithGetScriptSHA256(t *testing.T) {
	g := NewScriptGeneratorForApp("test-ns")
	cm := g.BuildConfigmap()
	sha := g.GetScriptSHA256()

	assert.Equal(t, sha, cm.Annotations[common.AnnotationCheckMountScriptSHA256])
}

func TestScriptGeneratorForApp_BuildConfigmap_NoDatasetLabel(t *testing.T) {
	// The app configmap is dataset-independent; it must NOT have LabelAnnotationDatasetId
	g := NewScriptGeneratorForApp("default")
	cm := g.BuildConfigmap()

	_, hasDatasetLabel := cm.Labels[common.LabelAnnotationDatasetId]
	assert.False(t, hasDatasetLabel, "app configmap should not have LabelAnnotationDatasetId")
	assert.Contains(t, cm.Annotations, common.AnnotationCheckMountScriptSHA256)
}

// --- scriptGeneratorHelper.RefreshConfigMapContents ---

func newTestHelperAndDataset(t *testing.T) (*scriptGeneratorHelper, *datav1alpha1.Dataset, types.NamespacedName) {
	t.Helper()
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			UID:       "test-uid",
		},
	}
	configMapKey := types.NamespacedName{Name: "test-cm", Namespace: "default"}
	helper := &scriptGeneratorHelper{
		configMapName:  "test-config",
		scriptFileName: "check-mount.sh",
		scriptContent:  "#!/bin/bash\necho hello",
		scriptSHA256:   computeScriptSHA256("#!/bin/bash\necho hello"),
	}
	return helper, dataset, configMapKey
}

func TestRefreshConfigMapContents_OverwritesData(t *testing.T) {
	helper, dataset, configMapKey := newTestHelperAndDataset(t)
	existing := helper.BuildConfigMap(dataset, configMapKey)
	existing.Data[helper.scriptFileName] = "old content"

	helper.RefreshConfigMapContents(dataset, configMapKey, existing)

	assert.Equal(t, "#!/bin/bash\necho hello", existing.Data[helper.scriptFileName])
}

func TestRefreshConfigMapContents_SetsSHA256Annotation(t *testing.T) {
	helper, dataset, configMapKey := newTestHelperAndDataset(t)
	existing := helper.BuildConfigMap(dataset, configMapKey)
	delete(existing.Annotations, common.AnnotationCheckMountScriptSHA256)

	helper.RefreshConfigMapContents(dataset, configMapKey, existing)

	assert.Equal(t, helper.scriptSHA256, existing.Annotations[common.AnnotationCheckMountScriptSHA256])
}

func TestRefreshConfigMapContents_OverwritesStaleSHA256(t *testing.T) {
	helper, dataset, configMapKey := newTestHelperAndDataset(t)
	existing := helper.BuildConfigMap(dataset, configMapKey)
	existing.Annotations[common.AnnotationCheckMountScriptSHA256] = "stale-sha"

	helper.RefreshConfigMapContents(dataset, configMapKey, existing)

	assert.Equal(t, helper.scriptSHA256, existing.Annotations[common.AnnotationCheckMountScriptSHA256])
}

func TestRefreshConfigMapContents_PreservesExtraLabels(t *testing.T) {
	helper, dataset, configMapKey := newTestHelperAndDataset(t)
	existing := helper.BuildConfigMap(dataset, configMapKey)
	existing.Labels["user-defined-label"] = "preserved"

	helper.RefreshConfigMapContents(dataset, configMapKey, existing)

	assert.Equal(t, "preserved", existing.Labels["user-defined-label"])
}

func TestRefreshConfigMapContents_PreservesExtraAnnotations(t *testing.T) {
	helper, dataset, configMapKey := newTestHelperAndDataset(t)
	existing := helper.BuildConfigMap(dataset, configMapKey)
	existing.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = "some-value"

	helper.RefreshConfigMapContents(dataset, configMapKey, existing)

	assert.Contains(t, existing.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

func TestRefreshConfigMapContents_InitializesNilLabels(t *testing.T) {
	helper, dataset, configMapKey := newTestHelperAndDataset(t)
	existing := helper.BuildConfigMap(dataset, configMapKey)
	existing.Labels = nil

	helper.RefreshConfigMapContents(dataset, configMapKey, existing)

	require.NotNil(t, existing.Labels)
	assert.Contains(t, existing.Labels, common.LabelAnnotationDatasetId)
}

func TestRefreshConfigMapContents_InitializesNilAnnotations(t *testing.T) {
	helper, dataset, configMapKey := newTestHelperAndDataset(t)
	existing := helper.BuildConfigMap(dataset, configMapKey)
	existing.Annotations = nil

	helper.RefreshConfigMapContents(dataset, configMapKey, existing)

	require.NotNil(t, existing.Annotations)
	assert.Contains(t, existing.Annotations, common.AnnotationCheckMountScriptSHA256)
}

func TestRefreshConfigMapContents_ReturnsSamePointer(t *testing.T) {
	helper, dataset, configMapKey := newTestHelperAndDataset(t)
	existing := helper.BuildConfigMap(dataset, configMapKey)
	result := helper.RefreshConfigMapContents(dataset, configMapKey, existing)

	assert.Same(t, existing, result)
}
