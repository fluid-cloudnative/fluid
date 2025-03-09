package engine

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func (t *CacheEngine) transformMasters(runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass, commonConfig *common.CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) error {
	value.Master = &common.CacheRuntimeComponentValue{
		Name:          t.getComponentName(common.ComponentTypeMaster),
		Namespace:     t.namespace,
		Enabled:       true,
		ComponentType: common.ComponentTypeMaster,
	}
	if runtimeClass.Topology.Master == nil || runtime.Spec.Master.Disabled {
		value.Master.Enabled = false
		return nil
	}
	if len(value.Master.Namespace) == 0 {
		value.Master.Namespace = "default"
	}
	if err := t.parseMasterFromRuntimeClass(runtimeClass, value); err != nil {
		return err
	}
	t.addCommonConfigForMaster(runtimeClass, commonConfig, value)

	t.parseMasterFromRuntime(runtime, value)

	return nil
}

func (e *CacheEngine) parseMasterFromRuntimeClass(runtimeClass *datav1alpha1.CacheRuntimeClass, value *common.CacheRuntimeValue) error {
	componentMaster := runtimeClass.Topology.Master
	value.Master.WorkloadType = componentMaster.WorkloadType
	value.Master.PodTemplateSpec = componentMaster.PodTemplateSpec

	if runtimeClass.Topology.Master.Service.Headless != nil {
		value.Master.Service = e.transformHeadlessServiceValue(value)
	}

	return nil
}

func (t *CacheEngine) parseMasterFromRuntime(runtime *datav1alpha1.CacheRuntime, value *common.CacheRuntimeValue) {
	podTemplateSpec := &value.Master.PodTemplateSpec

	// 1. image
	if len(runtime.Spec.Master.RuntimeVersion.Image) != 0 {
		podTemplateSpec.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", runtime.Spec.Master.RuntimeVersion.Image, runtime.Spec.Master.RuntimeVersion.ImageTag)
	}
	if len(runtime.Spec.Master.RuntimeVersion.ImagePullPolicy) != 0 {
		podTemplateSpec.Spec.Containers[0].ImagePullPolicy = corev1.PullPolicy(runtime.Spec.Master.RuntimeVersion.ImagePullPolicy)
	}
	if len(runtime.Spec.ImagePullSecrets) != 0 {
		podTemplateSpec.Spec.ImagePullSecrets = runtime.Spec.ImagePullSecrets
	}

	// 2. env
	if len(runtime.Spec.Master.Env) != 0 {
		podTemplateSpec.Spec.Containers[0].Env = append(value.Master.PodTemplateSpec.Spec.Containers[0].Env, runtime.Spec.Master.Env...)
	}
	podTemplateSpec.Spec.Containers[0].Env = append(podTemplateSpec.Spec.Containers[0].Env, t.generateCommonEnvs(runtime, value.Master.ComponentType)...)

	// 3. nodeSelector
	if len(runtime.Spec.Master.NodeSelector) != 0 {
		podTemplateSpec.Spec.NodeSelector = runtime.Spec.Master.NodeSelector
	}

	// 4. volume
	if len(runtime.Spec.Volumes) > 0 {
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, runtime.Spec.Volumes...)
	}

	if len(runtime.Spec.Master.VolumeMounts) > 0 {
		podTemplateSpec.Spec.Containers[0].VolumeMounts = append(podTemplateSpec.Spec.Containers[0].VolumeMounts, runtime.Spec.Master.VolumeMounts...)
	}

	// 5. metadate
	if len(runtime.Spec.PodMetadata.Annotations) != 0 {
		podTemplateSpec.Annotations = utils.UnionMapsWithOverride(value.Master.PodTemplateSpec.Annotations, runtime.Spec.PodMetadata.Annotations)
	}
	if len(runtime.Spec.PodMetadata.Labels) != 0 {
		podTemplateSpec.Labels = utils.UnionMapsWithOverride(value.Master.PodTemplateSpec.Labels, runtime.Spec.PodMetadata.Labels)
	}

	value.Master.Replicas = runtime.Spec.Master.Replicas
	if len(runtime.Spec.Master.Options) > 0 {
		value.Master.Options = utils.UnionMapsWithOverride(value.Master.Options, runtime.Spec.Master.Options)
	}

}

func (t *CacheEngine) addCommonConfigForMaster(runtimeClass *datav1alpha1.CacheRuntimeClass, commonConfig *common.CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) {
	componentMaster := value.Master

	componentMaster.PodTemplateSpec.Spec.ImagePullSecrets = append(
		componentMaster.PodTemplateSpec.Spec.ImagePullSecrets, commonConfig.ImagePullSecrets...)

	componentMaster.PodTemplateSpec.Spec.NodeSelector = utils.UnionMapsWithOverride(
		componentMaster.PodTemplateSpec.Spec.NodeSelector, commonConfig.NodeSelector)

	componentMaster.PodTemplateSpec.Spec.Tolerations = append(
		componentMaster.PodTemplateSpec.Spec.Tolerations, commonConfig.Tolerations...)

	componentMaster.PodTemplateSpec.Spec.Containers[0].Env = append(
		componentMaster.PodTemplateSpec.Spec.Containers[0].Env, commonConfig.Envs...)

	componentMaster.Owner = commonConfig.Owner
	componentMaster.Options = utils.UnionMapsWithOverride(runtimeClass.Topology.Master.Options, commonConfig.Options)

	if runtimeClass.Topology.Master.Dependencies.EncryptOption != nil {
		for _, encryptOptionVolume := range commonConfig.EncryptOptionConfigs.EncryptOptionVolumes {
			value.Master.PodTemplateSpec.Spec.Volumes = append(value.Master.PodTemplateSpec.Spec.Volumes, encryptOptionVolume)
		}

		for _, encryptOptionVolumeMount := range commonConfig.EncryptOptionConfigs.EncryptOptionVolumeMounts {
			for i := range value.Master.PodTemplateSpec.Spec.Containers {
				value.Master.PodTemplateSpec.Spec.Containers[i].VolumeMounts = append(value.Master.PodTemplateSpec.Spec.Containers[i].VolumeMounts, encryptOptionVolumeMount)
			}
		}
		value.Master.EncryptOption = commonConfig.EncryptOptionConfigs.EncryptOptionConfig
	}
	value.Master.PodTemplateSpec.Spec.Volumes = append(value.Master.PodTemplateSpec.Spec.Volumes, commonConfig.RuntimeConfigConfig.RuntimeConfigVolume)
	value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts = append(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts, commonConfig.RuntimeConfigConfig.RuntimeConfigVolumeMount)
}
