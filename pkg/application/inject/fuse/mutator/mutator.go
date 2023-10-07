package mutator

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FluidObjectMutator interface {
	Mutate() error
}

type MutatingContext struct {
	pvcName     string
	runtimeInfo base.RuntimeInfoInterface
	template    *common.FuseInjectionTemplate
	options     common.FuseSidecarInjectOption
	nameSuffix  string

	appendedVolumeNames         map[string]string
	datasetUsedInContainers     bool
	datasetUsedInInitContainers bool

	specs *MutatingPodSpecs
}

type MutatingPodSpecs struct {
	Volumes        []corev1.Volume
	VolumeMounts   []corev1.VolumeMount
	Containers     []corev1.Container
	InitContainers []corev1.Container
	MetaObj        metav1.ObjectMeta
}
