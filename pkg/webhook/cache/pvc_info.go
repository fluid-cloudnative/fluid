package cache

import (
	"sync"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"

	utilpointer "k8s.io/utils/pointer"
)

type PersistentVolumeClaimCachedInfo struct {
	// The uuid is used to check if the pvcCache is validate.
	// It's pvc's uuid
	cachedPVC *corev1.PersistentVolumeClaim

	// Check if the pvc belongs to the fluid dataset
	isBelongToDataset *bool

	// The runtime Info to cache
	runtimeInfo base.RuntimeInfoInterface

	// The fuse template to inject
	fuseTemplateToInject *common.FuseInjectionTemplate

	mu sync.RWMutex
}

func BuildPersistentVolumeClaimCachedInfo(cachedPVC *corev1.PersistentVolumeClaim,
	isBelongToDataset bool,
	runtimeInfo base.RuntimeInfoInterface,
	fuseTemplateToInject *common.FuseInjectionTemplate) *PersistentVolumeClaimCachedInfo {
	return &PersistentVolumeClaimCachedInfo{
		cachedPVC:            cachedPVC,
		isBelongToDataset:    utilpointer.BoolPtr(isBelongToDataset),
		runtimeInfo:          runtimeInfo,
		fuseTemplateToInject: fuseTemplateToInject,
	}
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
	return *p.isBelongToDataset
}

func (p *PersistentVolumeClaimCachedInfo) SetBelongToDataset(isBelongToDataset bool) {
	defer p.mu.Unlock()
	p.mu.Lock()
	p.isBelongToDataset = utilpointer.BoolPtr(isBelongToDataset)
}

func (p *PersistentVolumeClaimCachedInfo) GetRuntimeInfo() base.RuntimeInfoInterface {
	defer p.mu.RUnlock()
	p.mu.RLock()
	return p.runtimeInfo
}

func (p *PersistentVolumeClaimCachedInfo) SetRuntimeInfo(runtimeInfo base.RuntimeInfoInterface) {
	defer p.mu.Unlock()
	p.mu.Lock()
	p.runtimeInfo = runtimeInfo
}

func (p *PersistentVolumeClaimCachedInfo) GetFuseTemplateToInject() *common.FuseInjectionTemplate {
	defer p.mu.RUnlock()
	p.mu.RLock()
	return p.fuseTemplateToInject
}

func (p *PersistentVolumeClaimCachedInfo) SetFuseTemplateToInject(fuseTemplateToInject *common.FuseInjectionTemplate) {
	defer p.mu.Unlock()
	p.mu.Lock()
	p.fuseTemplateToInject = fuseTemplateToInject
}
