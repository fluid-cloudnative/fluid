package fileprefetcher

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestInjectFilePrefetcherSidecar_WithMatchingContainers_ShouldInsertAfterLastMatch(t *testing.T) {
	oldContainers := []corev1.Container{
		{Name: common.FuseContainerName + "-2"},
		{Name: common.FuseContainerName + "-1"},
		{Name: "other-container"},
		{Name: "another-container"},
	}
	prefetcher := &FilePrefetcher{}
	filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

	newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

	assert.Equal(t, 5, len(newContainers))
	assert.Equal(t, common.FuseContainerName+"-2", newContainers[0].Name)
	assert.Equal(t, common.FuseContainerName+"-1", newContainers[1].Name)
	assert.Equal(t, "file-prefetcher-ctr", newContainers[2].Name)
	assert.Equal(t, "other-container", newContainers[3].Name)
	assert.Equal(t, "another-container", newContainers[4].Name)
}

func TestInjectFilePrefetcherSidecar_WithoutMatchingContainers_ShouldInsertAtBeginning(t *testing.T) {
	oldContainers := []corev1.Container{
		{Name: "other-container"},
		{Name: "another-container"},
	}
	prefetcher := &FilePrefetcher{}
	filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

	newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

	assert.Equal(t, 3, len(newContainers))
	assert.Equal(t, "file-prefetcher-ctr", newContainers[0].Name)
	assert.Equal(t, "other-container", newContainers[1].Name)
	assert.Equal(t, "another-container", newContainers[2].Name)
}

func TestInjectFilePrefetcherSidecar_EmptyContainerList_ShouldOnlyContainFilePrefetcher(t *testing.T) {
	oldContainers := []corev1.Container{}
	filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

	prefetcher := &FilePrefetcher{}
	newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

	assert.Equal(t, 1, len(newContainers))
	assert.Equal(t, "file-prefetcher-ctr", newContainers[0].Name)
}
