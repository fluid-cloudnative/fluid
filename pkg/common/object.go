package common

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// The Application which is concerned by Fluid
type Application interface {
	GetObject() (obj runtime.Object)

	SetContainers(containers []corev1.Container, fields ...string)

	SetVolumes(volumes []corev1.Volume, fields ...string)

	GetVolumes(fields ...string) (volumes []corev1.Volume)

	GetContainers(fields ...string) (containers []corev1.Container)

	LocateContainers() (anchors []Anchor, err error)

	// LocateVolumes locates the volumes spec
	LocateVolumes() (anchors []Anchor, err error)

	// LocatePodSpec locates the pod spec or similar part in the CRD spec
	LocatePodSpec() (anchors []Anchor, err error)
}

type Anchor interface {
	Key() (id string)

	Path() (paths []string)
}
