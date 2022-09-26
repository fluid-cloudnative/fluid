package fuse

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func (s *Injector) injectCheckMountReadyScript(pod common.FluidObject, runtimeInfos map[string]base.RuntimeInfoInterface) error {
	objMeta, err := pod.GetMetaObject()
	if err != nil {
		return err
	}

	if !utils.FuseSidecarUnprivileged(objMeta.Labels) {
		// Skip if no need to check mount point ready
		return nil
	}

	appScriptGenerator, err := s.ensureScriptConfigMapExists(objMeta.Namespace)
	if err != nil {
		return err
	}

	volumes, err := pod.GetVolumes()
	if err != nil {
		return err
	}

	containers, err := pod.GetContainers()
	if err != nil {
		return err
	}

	for i := range containers {
		path2RuntimeTypeMap := collectDatasetVolumeMountInfo(containers[i].VolumeMounts, volumes, runtimeInfos)
		if len(path2RuntimeTypeMap) == 0 {
			continue
		}

		// todo: resolving name conflicts
		containers[i].VolumeMounts = append(containers[i].VolumeMounts, appScriptGenerator.GetVolumeMount())
		if utils.AppContainerPostStartInjectEnabled(objMeta.Labels) {
			if containers[i].Lifecycle != nil && containers[i].Lifecycle.PostStart != nil {
				//todo log
			} else {
				if containers[i].Lifecycle == nil {
					containers[i].Lifecycle = &corev1.Lifecycle{}
				}

				mountPaths, mountTypes := assembleMountInfos(path2RuntimeTypeMap)
				containers[i].Lifecycle.PostStart = appScriptGenerator.GetPostStartCommand(mountPaths, mountTypes)
			}
		}
	}

	err = pod.SetContainers(containers)
	if err != nil {
		return err
	}

	err = pod.SetVolumes(volumes)
	if err != nil {
		return err
	}

	return nil
}

func (s *Injector) ensureScriptConfigMapExists(namespace string) (*poststart.ScriptGeneratorForApp, error) {
	appScriptGen := poststart.NewScriptGeneratorForApp(namespace)

	cm := appScriptGen.BuildConfigmap()
	cmFound, err := kubeclient.IsConfigMapExist(s.client, cm.Name, cm.Namespace)
	if err != nil {
		return nil, err
	}

	if !cmFound {
		err = s.client.Create(context.TODO(), cm)
		if err != nil {
			if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
				return nil, err
			}
		}
	}

	return appScriptGen, nil
}

func collectDatasetVolumeMountInfo(volMounts []corev1.VolumeMount, volumes []corev1.Volume, runtimeInfos map[string]base.RuntimeInfoInterface) map[string]string {
	path2RuntimeTypeMap := map[string]string{}
	for _, volMount := range volMounts {
		vol := utils.FindVolumeByVolumeMount(volMount, volumes)
		if vol == nil {
			// todo: log
			continue
		}

		if vol.PersistentVolumeClaim != nil {
			if ri, ok := runtimeInfos[vol.PersistentVolumeClaim.ClaimName]; ok {
				path2RuntimeTypeMap[volMount.MountPath] = ri.GetRuntimeType()
			}
		}
	}

	return path2RuntimeTypeMap
}

func assembleMountInfos(path2RuntimeTypeMap map[string]string) (mountPathStr, mountTypeStr string) {
	var (
		mountPaths []string
		mountTypes []string
	)

	for k, v := range path2RuntimeTypeMap {
		mountPaths = append(mountPaths, k)
		mountTypes = append(mountTypes, v)
	}

	mountPathStr = strings.Join(mountPaths, ":")
	mountTypeStr = strings.Join(mountTypes, ":")

	return
}
