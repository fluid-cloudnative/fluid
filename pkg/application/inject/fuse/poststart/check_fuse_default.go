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
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// const (
// 	configMapName             = "check-mount"
// 	unprivilegedConfigMapName = configMapName + "-unprivileged"
// 	scriptName                = configMapName + ".sh"
// 	scriptPath                = "/" + scriptName
// )

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
)

// DefaultMountCheckScriptGenerator is a generator to render resources and specs related to post start mount-check script for the DefaultMutator
type defaultPostStartScriptGenerator struct {
	scriptGeneratorHelper
}

func NewDefaultPostStartScriptGenerator() *defaultPostStartScriptGenerator {
	return &defaultPostStartScriptGenerator{
		scriptGeneratorHelper: scriptGeneratorHelper{
			configMapName:   "check-mount",
			scriptFileName:  "check-mount.sh",
			scriptMountPath: "/check-mount.sh",
			scriptContent:   replacer.Replace(contentPrivilegedSidecar),
		},
	}
}

func (g *defaultPostStartScriptGenerator) GetPostStartCommand(mountPath, mountType, subPath string) (handler *corev1.LifecycleHandler) {
	// https://github.com/kubernetes/kubernetes/issues/25766
	cmd := []string{"bash", "-c", fmt.Sprintf("time %s %s %s %s >> /proc/1/fd/1", g.scriptMountPath, mountPath, mountType, subPath)}

	return &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{Command: cmd},
	}
}
