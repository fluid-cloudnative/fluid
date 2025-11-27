package utils

import (
	corev1 "k8s.io/api/core/v1"
)

// getContainerIndex returns the index of the container with the given name in the container list
// Returns -1 if the container is not found
func GetContainerIndex(containers []corev1.Container, containerName string) int {
	for i, container := range containers {
		if container.Name == containerName {
			return i
		}
	}
	return -1
}
