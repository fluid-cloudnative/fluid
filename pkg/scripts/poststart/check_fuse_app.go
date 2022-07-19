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

package poststart

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/utils/pointer"
)

const (
	appScriptName = "check-dataset-ready.sh"
	appScriptPath = "/" + appScriptName
	appVolName    = "check-dataset-ready"
)

type ScriptGeneratorForApp struct {
	name      string
	namespace string
	mountType string

	enablePostStartInjection bool
}

func NewScriptGeneratorForApp(namespacedKey types.NamespacedName, mountType string, enablePostStartInjection bool) *ScriptGeneratorForApp {
	return &ScriptGeneratorForApp{
		name:                     namespacedKey.Name,
		namespace:                namespacedKey.Namespace,
		mountType:                mountType,
		enablePostStartInjection: enablePostStartInjection,
	}
}

func (a *ScriptGeneratorForApp) BuildConfigmap(ownerReference metav1.OwnerReference) *corev1.ConfigMap {
	data := map[string]string{}
	content := contentPrivilegedSidecar

	data[appScriptName] = replacer.Replace(content)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            a.getConfigmapName(),
			Namespace:       a.namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: data,
	}
}

func (a *ScriptGeneratorForApp) getConfigmapName() string {
	return a.name + "-" + a.mountType + "-app-" + configMapName
}

func (a *ScriptGeneratorForApp) GetPostStartCommand(mountPath string) (handler *corev1.LifecycleHandler) {
	// Return non-null post start command only when PostStartInjeciton is enabled
	if a.enablePostStartInjection {
		// https://github.com/kubernetes/kubernetes/issues/25766
		cmd := []string{"bash", "-c", fmt.Sprintf("time %s %s %s >> /proc/1/fd/1", appScriptPath, mountPath, a.mountType)}
		handler = &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{Command: cmd},
		}
	}

	return
}

func (a *ScriptGeneratorForApp) GetVolume() (v corev1.Volume) {
	volName := appVolName
	var mode int32 = 0755
	return corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: a.getConfigmapName(),
				},
				DefaultMode: utilpointer.Int32Ptr(mode),
			},
		},
	}
}

func (a *ScriptGeneratorForApp) GetVolumeMount() (vm corev1.VolumeMount) {
	volName := appVolName
	return corev1.VolumeMount{
		Name:      volName,
		MountPath: appScriptPath,
		SubPath:   appScriptName,
		ReadOnly:  true,
	}
}
