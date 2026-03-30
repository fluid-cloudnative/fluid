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

package poststart

import (
	"crypto/sha256"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

type scriptGeneratorHelper struct {
	configMapName   string
	scriptContent   string
	scriptFileName  string
	scriptMountPath string
	scriptSHA256    string // first 63 chars of SHA256 of the scriptContent
}

// ScriptGenerator is the interface that concrete post-start script generators must implement.
// It is used by mutator helpers to create and refresh the check-mount ConfigMap.
type ScriptGenerator interface {
	BuildConfigMap(dataset *datav1alpha1.Dataset, configMapKey types.NamespacedName) *corev1.ConfigMap
	RefreshConfigMapContents(dataset *datav1alpha1.Dataset, configMapKey types.NamespacedName, existingCM *corev1.ConfigMap) *corev1.ConfigMap
	GetScriptSHA256() string
	GetNamespacedConfigMapKey(datasetKey types.NamespacedName, runtimeType string) types.NamespacedName
	GetVolume(configMapKey types.NamespacedName) corev1.Volume
	GetVolumeMount() corev1.VolumeMount
	GetPostStartCommand(mountPath, mountType, subPath string) *corev1.LifecycleHandler
}

// computeScriptSHA256 computes the SHA256 of content and returns the first 63 hex chars
// (K8s label values must be <= 63 characters; SHA256 hex is 64 chars).
func computeScriptSHA256(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum)[:63]
}

func (helper *scriptGeneratorHelper) BuildConfigMap(dataset *datav1alpha1.Dataset, configMapKey types.NamespacedName) *corev1.ConfigMap {
	data := map[string]string{}
	data[helper.scriptFileName] = helper.scriptContent
	// data[helper.scriptFileName] = replacer.Replace(helper.scriptContent)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapKey.Name,
			Namespace: configMapKey.Namespace,
			Labels: map[string]string{
				common.LabelAnnotationDatasetId: utils.GetDatasetId(configMapKey.Namespace, dataset.Name, string(dataset.UID)),
			},
			Annotations: map[string]string{
				common.AnnotationCheckMountScriptSHA256: helper.scriptSHA256,
			},
		},
		Data: data,
	}
}

// GetScriptSHA256 returns the SHA256 of the helper's script content.
// If the SHA was not set at construction time, it is computed lazily from scriptContent.
func (helper *scriptGeneratorHelper) GetScriptSHA256() string {
	if helper.scriptSHA256 != "" {
		return helper.scriptSHA256
	}
	if helper.scriptContent == "" {
		return ""
	}
	return computeScriptSHA256(helper.scriptContent)
}

// RefreshConfigMapContents updates existingCM's Data, Labels, and Annotations in-place to
// match what BuildConfigMap would produce for the given dataset and configMapKey, then returns
// the updated object. The caller is responsible for persisting the change.
func (helper *scriptGeneratorHelper) RefreshConfigMapContents(dataset *datav1alpha1.Dataset, configMapKey types.NamespacedName, existingCM *corev1.ConfigMap) *corev1.ConfigMap {
	newCM := helper.BuildConfigMap(dataset, configMapKey)
	existingCM.Data = newCM.Data
	if existingCM.Labels == nil {
		existingCM.Labels = map[string]string{}
	}
	for k, v := range newCM.Labels {
		existingCM.Labels[k] = v
	}
	if existingCM.Annotations == nil {
		existingCM.Annotations = map[string]string{}
	}
	for k, v := range newCM.Annotations {
		existingCM.Annotations[k] = v
	}
	return existingCM
}

func (helper *scriptGeneratorHelper) GetNamespacedConfigMapKey(datasetKey types.NamespacedName, runtimeType string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: datasetKey.Namespace,
		Name:      strings.ToLower(runtimeType) + "-" + helper.configMapName,
	}
}

func (helper *scriptGeneratorHelper) GetVolume(configMapKey types.NamespacedName) (v corev1.Volume) {
	var mode int32 = 0755
	return corev1.Volume{
		Name: helper.configMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapKey.Name,
				},
				DefaultMode: ptr.To(mode),
			},
		},
	}
}

func (helper *scriptGeneratorHelper) GetVolumeMount() (vm corev1.VolumeMount) {
	return corev1.VolumeMount{
		Name:      helper.configMapName,
		MountPath: helper.scriptMountPath,
		SubPath:   helper.scriptFileName,
		ReadOnly:  true,
	}
}
