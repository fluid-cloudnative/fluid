/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package poststart

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
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
