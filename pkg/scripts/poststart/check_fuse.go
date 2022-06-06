/*
Copyright 2021 The Fluid Authors.

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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/utils/pointer"
)

const (
	configMapName = "check-mount"
	scriptName    = configMapName + ".sh"
	scriptPath    = "/" + scriptName
)

var (
	replacer                 = strings.NewReplacer("¬", "`")
	contentPrivilegedSidecar = `#!/bin/bash

set -ex

ConditionPathIsMountPoint="$1"
MountType="$2"

count=0
# while ! mount | grep alluxio | grep  $ConditionPathIsMountPoint | grep -v grep
while ! mount | grep $ConditionPathIsMountPoint | grep $MountType
do
    sleep 3
    count=¬expr $count + 1¬
    if test $count -eq 10
    then
        echo "timed out!"
        exit 1
    fi
done

echo "succeed in checking mount point $ConditionPathIsMountPoint"
`
	contentUnprivilegedSidecar = `#!/bin/bash
set -ex

echo "Sending deivce ioctl to /dev/fuse"
chmod u+x /tools/ioctl_sync
/tools/ioctl_sync
`
)

type ScriptGeneratorForFuse struct {
	name      string
	namespace string
	mountPath string
	mountType string

	option common.FuseSidecarInjectOptions
}

func NewGenerator(namespacedKey types.NamespacedName, mountPath string, mountType string, option common.FuseSidecarInjectOptions) *ScriptGeneratorForFuse {
	return &ScriptGeneratorForFuse{
		name:      namespacedKey.Name,
		namespace: namespacedKey.Namespace,
		mountPath: mountPath,
		mountType: mountType,
		option:    option,
	}
}

func (f *ScriptGeneratorForFuse) BuildConfigmap(ownerReference metav1.OwnerReference) *corev1.ConfigMap {
	data := map[string]string{}
	var content string
	if f.option.EnableUnprivilegedSidecar {
		content = contentUnprivilegedSidecar
	} else {
		content = contentPrivilegedSidecar
	}
	data[scriptName] = replacer.Replace(content)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            f.getConfigmapName(),
			Namespace:       f.namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: data,
	}
}

func (f *ScriptGeneratorForFuse) getConfigmapName() string {
	return f.name + "-" + strings.ToLower(f.mountType) + "-" + configMapName
}

func (f *ScriptGeneratorForFuse) GetPostStartCommand() (handler *corev1.LifecycleHandler) {
	var cmd []string
	if f.option.EnableUnprivilegedSidecar {
		cmd = []string{"bash", "-c", fmt.Sprintf("time %s >> /proc/1/fd/1", scriptPath)}
	} else {
		// https://github.com/kubernetes/kubernetes/issues/25766
		cmd = []string{"bash", "-c", fmt.Sprintf("time %s %s %s >> /proc/1/fd/1", scriptPath, f.mountPath, f.mountType)}
	}
	handler = &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{Command: cmd},
	}
	return
}

func (f *ScriptGeneratorForFuse) GetVolume() (v corev1.Volume) {
	var mode int32 = 0755
	return corev1.Volume{
		Name: configMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: f.getConfigmapName(),
				},
				DefaultMode: utilpointer.Int32Ptr(mode),
			},
		},
	}
}

func (f *ScriptGeneratorForFuse) GetVolumeMount() (vm corev1.VolumeMount) {
	return corev1.VolumeMount{
		Name:      configMapName,
		MountPath: scriptPath,
		SubPath:   scriptName,
		ReadOnly:  true,
	}
}
