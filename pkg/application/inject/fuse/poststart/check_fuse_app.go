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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	appScriptName    = "check-fluid-mount-ready.sh"
	appScriptPath    = "/" + appScriptName
	appVolName       = "check-fluid-mount-ready"
	appConfigMapName = appVolName
)

var contentCheckMountReadyScript = `#!/bin/bash

set -e

if [[ "$#" -ne 2 ]]; then
  echo -e "Usage:"
  echo -e "  check-fluid-mount-ready.sh <path>[:<path>...] <runtime_type>[:<runtime_type>...]"
  echo -e "  (e.g. check-fluid-mount-ready.sh /data1:/data2:/data3 alluxio:jindo:juicefs)"
  exit 1
fi

mountPaths=( $(echo "$1" | tr ":" " ") )
mountTypes=( $(echo "$2" | tr ":" " ") )

if [[ "${#mountPaths[@]}" -ne "${#mountTypes[@]}" ]]; then
  echo "Error: length of mount paths must be equal to length of runtime types"
  exit 1
fi

for idx in "${!mountPaths[@]}"; do
	mp=${mountPaths[$idx]}
    mt=${mountTypes[$idx]}
	fstype="$mt"
	if [[ $mt == "jindo" ]]; then
	  fstype="fuse"
	fi
	count=0
	# Check fstype (9th item in /proc/self/mountinfo)
	while ! cat /proc/self/mountinfo | grep $mp | awk '{print $9}' | grep $fstype
	do
    	sleep 1
    	count=¬expr $count + 1¬
    	if test $count -eq 10
    	then
    	    echo "fail to check mount point $mp with runtimeType $mt: timed out for 10 seconds"
    	    exit 1
    	fi
	done
	echo "succeed in checking mount point $mp with runtimeType $mt"
done

`

type ScriptGeneratorForApp struct {
	namespace string
}

func NewScriptGeneratorForApp(namespace string) *ScriptGeneratorForApp {
	return &ScriptGeneratorForApp{
		namespace: namespace,
	}
}

func (a *ScriptGeneratorForApp) BuildConfigmap() *corev1.ConfigMap {
	data := map[string]string{}
	content := contentCheckMountReadyScript

	data[appScriptName] = replacer.Replace(content)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.getConfigmapName(),
			Namespace: a.namespace,
		},
		Data: data,
	}
}

func (a *ScriptGeneratorForApp) getConfigmapName() string {
	return appConfigMapName
}

func (a *ScriptGeneratorForApp) GetPostStartCommand(mountPaths string, mountTypes string) (handler *corev1.LifecycleHandler) {
	// Return non-null post start command only when PostStartInjeciton is enabled
	// https://github.com/kubernetes/kubernetes/issues/25766
	cmd := []string{"bash", "-c", fmt.Sprintf("time %s %s %s >> /proc/1/fd/1", appScriptPath, mountPaths, mountTypes)}
	handler = &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{Command: cmd},
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
				DefaultMode: ptr.To(mode),
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
