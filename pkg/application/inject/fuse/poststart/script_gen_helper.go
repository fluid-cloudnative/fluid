package poststart

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/utils/pointer"
)

type scriptGeneratorHelper struct {
	configMapName   string
	scriptContent   string
	scriptFileName  string
	scriptMountPath string
}

func (helper *scriptGeneratorHelper) BuildConfigMap(ownerReference metav1.OwnerReference, configMapKey types.NamespacedName) *corev1.ConfigMap {
	data := map[string]string{}
	data[helper.scriptFileName] = helper.scriptContent
	// data[helper.scriptFileName] = replacer.Replace(helper.scriptContent)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            configMapKey.Name,
			Namespace:       configMapKey.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: data,
	}
}

func (helper *scriptGeneratorHelper) GetConfigMapKeyByOwner(datasetKey types.NamespacedName, runtimeType string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: datasetKey.Namespace,
		Name:      datasetKey.Name + "-" + strings.ToLower(runtimeType) + "-" + helper.configMapName,
	}
}

func (helper *scriptGeneratorHelper) GetVolume(configMapKey types.NamespacedName) (v corev1.Volume) {
	var mode int32 = 0755
	return corev1.Volume{
		Name: helper.configMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapKey.Name,
				},
				DefaultMode: utilpointer.Int32Ptr(mode),
			},
		},
	}
}

func (helper *scriptGeneratorHelper) GetVolumeMount() (vm corev1.VolumeMount) {
	return corev1.VolumeMount{
		Name:      helper.configMapName,
		MountPath: helper.scriptMountPath,
		SubPath:   helper.scriptFileName,
		ReadOnly:  true,
	}
}
