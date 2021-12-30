package common

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Application interface {
	GetObject() (obj *unstructured.Unstructured)

	SetContainers(containers []corev1.Container, fields ...string)

	SetVolumes(volumes []corev1.Volume, fields ...string)

	GetVolumes(fields ...string) (volumes []corev1.Volume)

	GetContainers(fields ...string) (containers []corev1.Container)

	LocateContainers() (anchors []Anchor)

	LocateVolumes() (anchors []Anchor)
}

type Anchor interface {
	Key() (id string)

	Path() (paths []string)
}
