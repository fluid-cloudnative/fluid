package fuse

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func mutateVolumeMounts(containers []corev1.Container, appScriptGenerator *poststart.ScriptGeneratorForApp, datasetVolumeNames []string) []corev1.Container {
	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer

	for ci, container := range containers {
		// Add volumeMounts only when appScriptGenerator is non-null, which means fuse sidecar injection in privileged mode.
		if appScriptGenerator != nil {
			containers[ci].VolumeMounts = append(containers[ci].VolumeMounts, appScriptGenerator.GetVolumeMount())
		}

		// Set HostToContainer to the dataset volume mount point
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(datasetVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer

				// Inject postStartHook only when appScriptGenerator is non-null, which means fuse sidecar injection in privileged mode.
				if appScriptGenerator != nil {
					if postStart := appScriptGenerator.GetPostStartCommand(volumeMount.MountPath); postStart != nil {
						if containers[ci].Lifecycle != nil && containers[ci].Lifecycle.PostStart != nil {
							warningStr := fmt.Sprintf("===> Skipping inject post start command because container %v already have one", containers[ci].Name)
							log.Info(warningStr)
							continue
						} else {
							if containers[ci].Lifecycle == nil {
								containers[ci].Lifecycle = &corev1.Lifecycle{}
							}
							containers[ci].Lifecycle.PostStart = appScriptGenerator.GetPostStartCommand(volumeMount.MountPath)
						}
					}

				}
			}
		}
	}

	return containers
}
