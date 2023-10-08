package mutator

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FluidObjectMutator interface {
	Mutate() (*MutatingPodSpecs, error)
}

// MutatingContext describes all the information required for mutating the specs
type MutatingContext struct {
	// info required for the mutation
	PvcName     string
	RuntimeInfo base.RuntimeInfoInterface
	Template    *common.FuseInjectionTemplate
	Options     common.FuseSidecarInjectOption
	NameSuffix  string

	// stateful variables shared across the mutation
	AppendedVolumeNames         map[string]string
	DatasetUsedInContainers     bool
	DatasetUsedInInitContainers bool

	Specs *MutatingPodSpecs
}

// MutatingPodSpecs stores all the mutable properties for a FluidObject like a Pod.
type MutatingPodSpecs struct {
	Volumes        []corev1.Volume
	VolumeMounts   []corev1.VolumeMount
	Containers     []corev1.Container
	InitContainers []corev1.Container
	MetaObj        metav1.ObjectMeta
}
