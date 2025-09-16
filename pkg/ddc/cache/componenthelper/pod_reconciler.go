package componenthelper

import (
	"fmt"
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/componenthelper/utils"
)

func podTemplateSpecEqual(template1, template2 corev1.PodTemplateSpec) (bool, error) {
	if equal, err := objectMetaEqual(template1.ObjectMeta, template2.ObjectMeta); !equal {
		return false, fmt.Errorf("objectMeta not equal: %s", err.Error())
	}

	if equal, err := podSpecEqual(template1.Spec, template2.Spec); !equal {
		return false, fmt.Errorf("spec not equal: %s", err.Error())
	}

	return true, nil
}

func objectMetaEqual(meta1, meta2 metav1.ObjectMeta) (bool, error) {
	meta1Copy := meta1.DeepCopy()
	meta2Copy := meta2.DeepCopy()

	meta1Copy.Labels = utils.FilterSystemLabels(meta1Copy.Labels)
	meta2Copy.Labels = utils.FilterSystemLabels(meta2Copy.Labels)
	if !mapsEqual(meta1Copy.Labels, meta2Copy.Labels) {
		return false, fmt.Errorf("label not equal, old [%s], new [%s]", meta1Copy.Labels, meta2Copy.Labels)
	}

	meta1Copy.Annotations = utils.FilterSystemAnnotations(meta1Copy.Annotations)
	meta2Copy.Annotations = utils.FilterSystemAnnotations(meta2Copy.Annotations)
	if !mapsEqual(meta1Copy.Annotations, meta2Copy.Annotations) {
		return false, fmt.Errorf("annotation not equal, old [%s], new [%s]", meta1Copy.Annotations, meta2Copy.Annotations)
	}
	return true, nil
}

// podSpecEqual 比较 PodTemplateSpec
func podSpecEqual(spec1, spec2 corev1.PodSpec) (bool, error) {
	if len(spec1.Containers) != len(spec2.Containers) {
		return false, fmt.Errorf("pod template spec containers len not equal")
	}

	// 对容器进行排序后比较
	containers1 := sortContainers(spec1.Containers)
	containers2 := sortContainers(spec2.Containers)

	for i := range containers1 {
		if equal, err := containerEqual(containers1[i], containers2[i]); !equal {
			return false, fmt.Errorf("container not equal: %s", err.Error())
		}
	}

	// 比较 volumes
	if equal, err := volumesEqual(spec1.Volumes, spec2.Volumes); !equal {
		return false, fmt.Errorf("podTemplate volumes not equal: %s", err.Error())
	}

	return true, nil
}

// containerEqual 比较容器
func containerEqual(c1, c2 corev1.Container) (bool, error) {
	if c1.Name != c2.Name {
		return false, fmt.Errorf("container name not equal")
	}

	if c1.Image != c2.Image {
		return false, fmt.Errorf("container image not equal")
	}

	if !reflect.DeepEqual(c1.Command, c2.Command) {
		return false, fmt.Errorf("container command not equal")
	}

	if !reflect.DeepEqual(c1.Args, c2.Args) {
		return false, fmt.Errorf("container args not equal")
	}

	if !reflect.DeepEqual(c1.Ports, c2.Ports) {
		return false, fmt.Errorf("container ports not equal")
	}

	if !reflect.DeepEqual(c1.Resources, c2.Resources) {
		return false, fmt.Errorf("container resources not equal")
	}

	if c1.ImagePullPolicy != "" && c2.ImagePullPolicy != "" && c1.ImagePullPolicy != c2.ImagePullPolicy {
		return false, fmt.Errorf("container image pull policy not equal, old: %s, new: %s", c1.ImagePullPolicy, c2.ImagePullPolicy)
	}

	// 比较环境变量
	if equal, err := envVarsEqual(c1.Env, c2.Env); !equal {
		return false, fmt.Errorf("env not equal: %s", err.Error())
	}

	// 比较挂载点
	if equal, err := volumeMountsEqual(c1.VolumeMounts, c2.VolumeMounts); !equal {
		return false, fmt.Errorf("podTemplate volumes mounts not equal: %s", err.Error())
	}

	return true, nil

}

// envVarsEqual 比较环境变量
func envVarsEqual(env1, env2 []corev1.EnvVar) (bool, error) {
	env1 = utils.FilterSystemEnvs(env1)
	env2 = utils.FilterSystemEnvs(env2)
	if len(env1) != len(env2) {
		return false, fmt.Errorf("env vars len not equal")
	}

	sortedEnv1 := make([]corev1.EnvVar, len(env1))
	sortedEnv2 := make([]corev1.EnvVar, len(env2))
	copy(sortedEnv1, env1)
	copy(sortedEnv2, env2)

	// 按名称排序
	sort.Slice(sortedEnv1, func(i, j int) bool {
		return sortedEnv1[i].Name < sortedEnv1[j].Name
	})
	sort.Slice(sortedEnv2, func(i, j int) bool {
		return sortedEnv2[i].Name < sortedEnv2[j].Name
	})

	for i := range sortedEnv1 {
		if !reflect.DeepEqual(sortedEnv1[i].Value, sortedEnv2[i].Value) {
			return false, fmt.Errorf("env vars %s value not equal, old: %v, new: %v", sortedEnv1[i].Name, sortedEnv1[i].Value, sortedEnv2[i].Value)
		}
		if !reflect.DeepEqual(sortedEnv1[i].Name, sortedEnv2[i].Name) {
			return false, fmt.Errorf("env vars name not equal")
		}
	}

	return true, nil
}

// volumesEqual 比较卷
func volumesEqual(vol1, vol2 []corev1.Volume) (bool, error) {
	if len(vol1) != len(vol2) {
		return false, fmt.Errorf("volumes not equal")
	}

	sortedVol1 := make([]corev1.Volume, len(vol1))
	sortedVol2 := make([]corev1.Volume, len(vol2))
	copy(sortedVol1, vol1)
	copy(sortedVol2, vol2)

	// 按名称排序
	sort.Slice(sortedVol1, func(i, j int) bool {
		return sortedVol1[i].Name < sortedVol1[j].Name
	})
	sort.Slice(sortedVol2, func(i, j int) bool {
		return sortedVol2[i].Name < sortedVol2[j].Name
	})

	// 比较volume名称是否一致
	for i := range sortedVol1 {
		if !reflect.DeepEqual(sortedVol1[i].Name, sortedVol2[i].Name) {
			return false, fmt.Errorf("volume name not equal")
		}
	}

	return true, nil
}

// volumeMountsEqual 比较卷挂载
func volumeMountsEqual(vm1, vm2 []corev1.VolumeMount) (bool, error) {
	if len(vm1) != len(vm2) {
		return false, fmt.Errorf("volume mounts len not equal")
	}

	sortedVM1 := make([]corev1.VolumeMount, len(vm1))
	sortedVM2 := make([]corev1.VolumeMount, len(vm2))
	copy(sortedVM1, vm1)
	copy(sortedVM2, vm2)

	// 按名称排序
	sort.Slice(sortedVM1, func(i, j int) bool {
		return sortedVM1[i].Name < sortedVM1[j].Name
	})
	sort.Slice(sortedVM2, func(i, j int) bool {
		return sortedVM2[i].Name < sortedVM2[j].Name
	})

	for i := range sortedVM1 {
		if !reflect.DeepEqual(sortedVM1[i].Name, sortedVM2[i].Name) {
			return false, fmt.Errorf("volume mount name not equal")
		}
	}

	return true, nil
}

// sortContainers 对容器按名称排序
func sortContainers(containers []corev1.Container) []corev1.Container {
	sorted := make([]corev1.Container, len(containers))
	copy(sorted, containers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// mapsEqual compares two map[string]string.
// It returns true if both maps are nil or empty.
// Otherwise, it compares keys and values for equality.
func mapsEqual(map1, map2 map[string]string) bool {
	isMap1Empty := map1 == nil || len(map1) == 0
	isMap2Empty := map2 == nil || len(map2) == 0

	if isMap1Empty && isMap2Empty {
		return true
	}

	if isMap1Empty != isMap2Empty {
		return false
	}

	if len(map1) != len(map2) {
		return false
	}

	for k, v := range map1 {
		if val2, ok := map2[k]; !ok || val2 != v {
			return false
		}
	}

	return true
}
