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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/utils/pointer"
)

type scriptGeneratorHelper struct {
	configMapName   string
	scriptContent   string
	scriptFileName  string
	scriptMountPath string
}

func (helper *scriptGeneratorHelper) BuildConfigMap(ownerReference metav1.OwnerReference, configMapKey types.NamespacedName) *corev1.ConfigMap {
	data := map[string]string{}
	data[helper.scriptFileName] = helper.scriptContent
	// data[helper.scriptFileName] = replacer.Replace(helper.scriptContent)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            configMapKey.Name,
			Namespace:       configMapKey.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: data,
	}
}

func (helper *scriptGeneratorHelper) GetConfigMapKeyByOwner(datasetKey types.NamespacedName, runtimeType string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: datasetKey.Namespace,
		Name:      datasetKey.Name + "-" + strings.ToLower(runtimeType) + "-" + helper.configMapName,
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
				DefaultMode: utilpointer.Int32Ptr(mode),
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
