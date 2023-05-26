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
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/utils/pointer"
)

const (
	configMapName             = "check-mount"
	unprivilegedConfigMapName = configMapName + "-unprivileged"
	scriptName                = configMapName + ".sh"
	scriptPath                = "/" + scriptName
)

var (
	replacer                 = strings.NewReplacer("¬", "`")
	contentPrivilegedSidecar = `#!/bin/bash

set -ex

ConditionPathIsMountPoint="$1"
MountType="$2"
SubPath="$3"

# grep /dev/fuse if the mountType equals to jindo
if [[ "$MountType" == "jindo" ]]; then
  MountType=/dev/fuse
fi

count=0
# while ! mount | grep alluxio | grep  $ConditionPathIsMountPoint | grep -v grep
while ! cat /proc/self/mountinfo | grep $ConditionPathIsMountPoint | grep $MountType
do
    sleep 3
    count=¬expr $count + 1¬
    if test $count -eq 10
    then
        echo "timed out!"
        exit 1
    fi
done

# different with csi, as here the mount point is the parent dir of the fuse mount point, 
if [ ! -e  $ConditionPathIsMountPoint/*/$SubPath ] ; then
  echo "sub path [$SubPath] not exist!"
  exit 2
fi

echo "succeed in checking mount point $ConditionPathIsMountPoint"
`
	contentUnprivilegedSidecar = `#!/bin/bash
set -ex

echo "Sending device ioctl to /dev/fuse"
/tools/ioctl_sync
echo "Device ioctl done. Post start script finished"
`
)

type ScriptGeneratorForFuse struct {
	name      string
	namespace string
	mountPath string
	mountType string
	subPath   string

	option common.FuseSidecarInjectOption
}

func NewGenerator(namespacedKey types.NamespacedName, mountPath string, mountType string, subPath string, option common.FuseSidecarInjectOption) *ScriptGeneratorForFuse {
	return &ScriptGeneratorForFuse{
		name:      namespacedKey.Name,
		namespace: namespacedKey.Namespace,
		mountPath: mountPath,
		mountType: mountType,
		subPath:   subPath,
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
	if f.option.EnableUnprivilegedSidecar {
		return f.name + "-" + strings.ToLower(f.mountType) + "-" + unprivilegedConfigMapName
	} else {
		return f.name + "-" + strings.ToLower(f.mountType) + "-" + configMapName
	}

}

func (f *ScriptGeneratorForFuse) GetPostStartCommand() (handler *corev1.LifecycleHandler) {
	var cmd []string
	if f.option.EnableUnprivilegedSidecar {
		cmd = []string{"bash", "-c", fmt.Sprintf("time %s >> /proc/1/fd/1", scriptPath)}
	} else {
		// https://github.com/kubernetes/kubernetes/issues/25766
		cmd = []string{"bash", "-c", fmt.Sprintf("time %s %s %s %s >> /proc/1/fd/1", scriptPath, f.mountPath, f.mountType, f.subPath)}
	}
	handler = &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{Command: cmd},
	}
	return
}

func (f *ScriptGeneratorForFuse) GetVolume() (v corev1.Volume) {
	var volName string
	if f.option.EnableUnprivilegedSidecar {
		volName = unprivilegedConfigMapName
	} else {
		volName = configMapName
	}
	var mode int32 = 0755
	return corev1.Volume{
		Name: volName,
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
	var volName string
	if f.option.EnableUnprivilegedSidecar {
		volName = unprivilegedConfigMapName
	} else {
		volName = configMapName
	}
	return corev1.VolumeMount{
		Name:      volName,
		MountPath: scriptPath,
		SubPath:   scriptName,
		ReadOnly:  true,
	}
}
