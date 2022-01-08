package common

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ApplicationPodSpec simulates the V1 Pod Spec, it has v1.volumes, v1.containers
type ApplicationPodSpec interface {
	GetVolumes() (volumes []corev1.Volume, err error)

	SetVolumes(volumes []corev1.Volume) (err error)

	GetContainers() (containers []corev1.Container, err error)

	SetContainers(containers []corev1.Container) (err error)
}

// The Application which is concerned by Fluid
type Application interface {
	GetPodSpec() (spec ApplicationPodSpec, err error)

	GetObject() (obj runtime.Object)

	SetContainers(containers []corev1.Container, fields ...string)

	SetVolumes(volumes []corev1.Volume, fields ...string)

	GetVolumes(fields ...string) (volumes []corev1.Volume)

	GetContainers(fields ...string) (containers []corev1.Container)

	LocateContainers() (anchors []Anchor, err error)

	// LocateVolumes locates the volumes spec
	LocateVolumes() (anchors []Anchor, err error)

	// LocatePodSpecs locates the pod spec or similar part in the CRD spec
	LocatePodSpecs() (anchors []Anchor, err error)

	LocateVolumeMounts() (anchors []Anchor, err error)
}

type Anchor interface {
	Key() (id string)

	Path() (paths []string)
}
