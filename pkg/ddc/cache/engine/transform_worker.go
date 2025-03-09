package engine

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *CacheEngine) transformWorkers(runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass, commonValue *common.CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) (err error) {
	value.Worker = &common.CacheRuntimeComponentValue{
		Name:          t.getComponentName(common.ComponentTypeWorker),
		Namespace:     t.namespace,
		Enabled:       true,
		ComponentType: common.ComponentTypeWorker,
	}
	if runtimeClass.Topology.Worker == nil || runtime.Spec.Worker.Disabled {
		value.Worker.Enabled = false
		return nil
	}
	if len(value.Worker.Namespace) == 0 {
		value.Worker.Namespace = "default"
	}
	if err := t.parseWorkerFromRuntimeClass(runtimeClass, value); err != nil {
		return err
	}

	t.addCommonConfigForWorker(runtimeClass, commonValue, value)

	if err := t.parseWorkerFromRuntime(runtime, value); err != nil {
		return err
	}

	return t.setWorkerAffinity(value)
}

func (e *CacheEngine) setWorkerAffinity(value *common.CacheRuntimeValue) error {
	var (
		name      = e.name
		namespace = e.namespace
	)

	if value.Worker.PodTemplateSpec.Spec.Affinity != nil {
		return nil
	}
	affinityToUpdate := &corev1.Affinity{}
	dataset, err := utils.GetDataset(e.Client, name, namespace)
	if err != nil {
		return err
	}
	if dataset.IsExclusiveMode() {
		affinityToUpdate.PodAntiAffinity = &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "fluid.io/dataset",
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		}
	} else {
		affinityToUpdate.PodAntiAffinity = &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					// The default weight is 50
					Weight: 50,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "fluid.io/dataset",
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "fluid.io/dataset-placement",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{string(datav1alpha1.ExclusiveMode)},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		}
	}

	// 3. Prefer to locate on the node which already has fuse on it
	if affinityToUpdate.NodeAffinity == nil {
		affinityToUpdate.NodeAffinity = &corev1.NodeAffinity{}
	}

	if len(affinityToUpdate.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		affinityToUpdate.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.PreferredSchedulingTerm{}
	}

	affinityToUpdate.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
		append(affinityToUpdate.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			corev1.PreferredSchedulingTerm{
				Weight: 100,
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      utils.GetFuseLabelName(e.namespace, e.name, string(dataset.UID)),
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			})

	// 3. set node affinity if possible
	if dataset.Spec.NodeAffinity != nil {
		if dataset.Spec.NodeAffinity.Required != nil {
			affinityToUpdate.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution =
				dataset.Spec.NodeAffinity.Required
		}
	}
	value.Worker.PodTemplateSpec.Spec.Affinity = affinityToUpdate

	return nil
}

func (e *CacheEngine) parseWorkerFromRuntimeClass(runtimeClass *datav1alpha1.CacheRuntimeClass, value *common.CacheRuntimeValue) error {
	componentWorker := runtimeClass.Topology.Worker
	value.Worker.WorkloadType = componentWorker.WorkloadType

	value.Worker.PodTemplateSpec = componentWorker.PodTemplateSpec

	if runtimeClass.Topology.Worker.Service.Headless != nil {
		value.Worker.Service = e.transformHeadlessServiceValue(value)
	}
	return nil
}

func (t *CacheEngine) parseWorkerFromRuntime(runtime *datav1alpha1.CacheRuntime, value *common.CacheRuntimeValue) error {
	podTemplateSpec := &value.Worker.PodTemplateSpec

	// 1. image
	if len(runtime.Spec.Worker.RuntimeVersion.Image) != 0 {
		podTemplateSpec.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", runtime.Spec.Worker.RuntimeVersion.Image, runtime.Spec.Worker.RuntimeVersion.ImageTag)
	}
	if len(runtime.Spec.Worker.RuntimeVersion.ImagePullPolicy) != 0 {
		podTemplateSpec.Spec.Containers[0].ImagePullPolicy = corev1.PullPolicy(runtime.Spec.Worker.RuntimeVersion.ImagePullPolicy)
	}
	if len(runtime.Spec.ImagePullSecrets) != 0 {
		podTemplateSpec.Spec.ImagePullSecrets = runtime.Spec.ImagePullSecrets
	}

	// 2. env
	if len(runtime.Spec.Worker.Env) != 0 {
		podTemplateSpec.Spec.Containers[0].Env = append(value.Worker.PodTemplateSpec.Spec.Containers[0].Env, runtime.Spec.Worker.Env...)
	}
	podTemplateSpec.Spec.Containers[0].Env = append(podTemplateSpec.Spec.Containers[0].Env, t.generateCommonEnvs(runtime, value.Worker.ComponentType)...)

	// 3. nodeSelector
	if len(runtime.Spec.Worker.NodeSelector) != 0 {
		podTemplateSpec.Spec.NodeSelector = runtime.Spec.Worker.NodeSelector
	}

	// 4. volume
	if len(runtime.Spec.Volumes) > 0 {
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, runtime.Spec.Volumes...)
	}

	if len(runtime.Spec.Worker.VolumeMounts) > 0 {
		podTemplateSpec.Spec.Containers[0].VolumeMounts = append(podTemplateSpec.Spec.Containers[0].VolumeMounts, runtime.Spec.Worker.VolumeMounts...)
	}

	// 5. metadate
	if len(runtime.Spec.PodMetadata.Annotations) != 0 {
		podTemplateSpec.Annotations = utils.UnionMapsWithOverride(value.Worker.PodTemplateSpec.Annotations, runtime.Spec.PodMetadata.Annotations)
	}
	if len(runtime.Spec.PodMetadata.Labels) != 0 {
		podTemplateSpec.Labels = utils.UnionMapsWithOverride(value.Worker.PodTemplateSpec.Labels, runtime.Spec.PodMetadata.Labels)
	}

	// 6. resources
	t.transformResourcesForContainer(runtime.Spec.Worker.Resources, &value.Worker.PodTemplateSpec.Spec.Containers[0])

	if len(runtime.Spec.Worker.TieredStore.Levels) > 0 {
		tieredStoreConfig, err := t.transformTieredStore(runtime.Spec.Worker.TieredStore.Levels)
		if err != nil {
			return err
		}
		podTemplateSpec.Spec.Containers[0].VolumeMounts = append(
			podTemplateSpec.Spec.Containers[0].VolumeMounts, tieredStoreConfig.CacheVolumeMounts...)
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, tieredStoreConfig.CacheVolumes...)
		value.Worker.TieredStore = tieredStoreConfig.CacheVolumeOptions
	}

	value.Worker.Replicas = runtime.Spec.Worker.Replicas
	if len(runtime.Spec.Worker.Options) > 0 {
		value.Worker.Options = utils.UnionMapsWithOverride(value.Worker.Options, runtime.Spec.Worker.Options)
	}

	return nil
}

func (t *CacheEngine) addCommonConfigForWorker(runtimeClass *datav1alpha1.CacheRuntimeClass, commonConfig *common.CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) {
	componentWorker := value.Worker

	componentWorker.PodTemplateSpec.Spec.ImagePullSecrets = append(
		componentWorker.PodTemplateSpec.Spec.ImagePullSecrets, commonConfig.ImagePullSecrets...)

	componentWorker.PodTemplateSpec.Spec.NodeSelector = utils.UnionMapsWithOverride(
		componentWorker.PodTemplateSpec.Spec.NodeSelector, commonConfig.NodeSelector)

	componentWorker.PodTemplateSpec.Spec.Tolerations = append(
		componentWorker.PodTemplateSpec.Spec.Tolerations, commonConfig.Tolerations...)

	componentWorker.PodTemplateSpec.Spec.Containers[0].Env = append(
		componentWorker.PodTemplateSpec.Spec.Containers[0].Env, commonConfig.Envs...)

	componentWorker.Owner = commonConfig.Owner
	componentWorker.Options = utils.UnionMapsWithOverride(runtimeClass.Topology.Worker.Options, commonConfig.Options)

	if runtimeClass.Topology.Worker.Dependencies.EncryptOption != nil {
		for _, encryptOptionVolume := range commonConfig.EncryptOptionConfigs.EncryptOptionVolumes {
			value.Worker.PodTemplateSpec.Spec.Volumes = append(value.Worker.PodTemplateSpec.Spec.Volumes, encryptOptionVolume)
		}

		for _, encryptOptionVolumeMount := range commonConfig.EncryptOptionConfigs.EncryptOptionVolumeMounts {
			for i := range value.Worker.PodTemplateSpec.Spec.Containers {
				value.Worker.PodTemplateSpec.Spec.Containers[i].VolumeMounts = append(value.Worker.PodTemplateSpec.Spec.Containers[i].VolumeMounts, encryptOptionVolumeMount)
			}
		}
		value.Master.EncryptOption = commonConfig.EncryptOptionConfigs.EncryptOptionConfig
	}
	value.Worker.PodTemplateSpec.Spec.Volumes = append(value.Worker.PodTemplateSpec.Spec.Volumes, commonConfig.RuntimeConfigConfig.RuntimeConfigVolume)
	value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts = append(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts, commonConfig.RuntimeConfigConfig.RuntimeConfigVolumeMount)

}
