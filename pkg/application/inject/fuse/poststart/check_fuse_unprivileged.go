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
)

var (
	contentUnprivilegedSidecar = `#!/bin/bash
set -ex

echo "Sending device ioctl to /dev/fuse"
/tools/ioctl_sync
echo "Device ioctl done. Post start script finished"
`
)

// unprivilegedPostStartScriptGenerator is a generator to render resources and specs related to post start mount-check script for the UnprivilegedMutator
type unprivilegedPostStartScriptGenerator struct {
	scriptGeneratorHelper
}

func NewUnprivilegedPostStartScriptGenerator() *unprivilegedPostStartScriptGenerator {
	return &unprivilegedPostStartScriptGenerator{
		scriptGeneratorHelper: scriptGeneratorHelper{
			configMapName:   "check-mount-unprivileged",
			scriptFileName:  "check-mount.sh",
			scriptMountPath: "/check-mount.sh",
			scriptContent:   replacer.Replace(contentUnprivilegedSidecar),
		},
	}
}

func (g *unprivilegedPostStartScriptGenerator) GetPostStartCommand() (handler *corev1.LifecycleHandler) {
	cmd := []string{"bash", "-c", fmt.Sprintf("time %s >> /proc/1/fd/1", g.scriptMountPath)}

	return &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{Command: cmd},
	}
}
