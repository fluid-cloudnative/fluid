package cache

import (
	"sync"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
)

type PersistentVolumeClaimCachedInfo struct {
	// The uuid is used to check if the pvcCache is validate.
	// It's pvc's uuid
	cachedPVC *corev1.PersistentVolumeClaim

	// Check if the pvc belongs to the fluid dataset
	isBelongToDataset bool

	// The runtime Info to cache
	runtimeInfo base.RuntimeInfoInterface

	// The fuse template to inject
	fuseTemplateToInject *common.FuseInjectionTemplate

	mu sync.RWMutex
}

func (p *PersistentVolumeClaimCachedInfo) GetCachedPVC() *corev1.PersistentVolumeClaim {
	defer p.mu.RUnlock()
	p.mu.RLock()
	return p.cachedPVC
}

func (p *PersistentVolumeClaimCachedInfo) SetCachedPVC(cachedPVC *corev1.PersistentVolumeClaim) {
	defer p.mu.Unlock()
	p.mu.Lock()
	p.cachedPVC = cachedPVC
}

func (p *PersistentVolumeClaimCachedInfo) IsBelongToDataset() bool {
	defer p.mu.RUnlock()
	p.mu.RLock()
	return p.isBelongToDataset
}

func (p *PersistentVolumeClaimCachedInfo) SetDataset(isBelongToDataset bool) {
	defer p.mu.Unlock()
	p.mu.Lock()
	p.isBelongToDataset = isBelongToDataset
}
