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
