package common

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Object simulates the V1 Pod Spec, it has v1.volumes, v1.containers inside
type Object interface {
	GetRoot() runtime.Object

	GetVolumesPtr() Pointer

	GetContainersPtr() Pointer

	GetVolumes() (volumes []corev1.Volume, err error)

	SetVolumes(volumes []corev1.Volume) (err error)

	GetContainers() (containers []corev1.Container, err error)

	SetContainers(containers []corev1.Container) (err error)
}

// The Application which is using Fluid,
// and it has serveral PodSpecs.
type Application interface {
	GetPodSpecs() (specs []Object, err error)

	SetPodSpecs(specs []Object) (err error)

	GetObject() (obj runtime.Object)

	// SetContainers(containers []corev1.Container, fields ...string)

	// SetVolumes(volumes []corev1.Volume, fields ...string)

	// GetVolumes(fields ...string) (volumes []corev1.Volume)

	// GetContainers(fields ...string) (containers []corev1.Container)

	LocateContainers() (ptrs []Pointer, err error)

	// LocateVolumes locates the volumes spec
	LocateVolumes() (ptrs []Pointer, err error)

	// LocatePodSpecs locates the pod spec or similar part in the CRD spec
	LocatePodSpecs() (ptrs []Pointer, err error)

	LocateVolumeMounts() (ptrs []Pointer, err error)
}

type Pointer interface {
	Key() (id string)

	Paths() (paths []string)

	Parent() (p Pointer, err error)

	Child(name string) (p Pointer)
}
