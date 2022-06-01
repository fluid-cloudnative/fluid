package mutating

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/cache"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CheckIfPVCIsDataset(pvc *corev1.PersistentVolumeClaim, log logr.Logger) (isDataset bool) {

	// cachedInfo, err := cache.GetOrCreateCachedInfo(pvc)
	// log.Info("Failed to get cache, but ignore it.", "err", err)

	return kubeclient.CheckIfPVCIsDataset(pvc)
}

func buildRuntimeInfo(client client.Client,
	pvc *corev1.PersistentVolumeClaim,
	log logr.Logger) (runtimeInfo base.RuntimeInfoInterface, err error) {
	cachedInfo, err := cache.GetOrCreateCachedInfo(pvc)
	if err != nil {
		log.Info("Failed to get cache, but ignore it.", "err", err)
	}

	// if no cache info is available
	if cachedInfo == nil {
		return buildRuntimeInfoInternal(client, pvc, log)
	}

	runtimeInfo = cachedInfo.GetRuntimeInfo()
	// runtime info is not set
	if runtimeInfo == nil {
		runtimeInfo, err = buildRuntimeInfoInternal(client, pvc, log)
		if err != nil {
			return
		}
		cachedInfo.SetRuntimeInfo(runtimeInfo)
	}

	return
}

func buildRuntimeInfoInternal(client client.Client,
	pvc *corev1.PersistentVolumeClaim,
	log logr.Logger) (runtimeInfo base.RuntimeInfoInterface, err error) {
	runtimeInfo, err = base.GetRuntimeInfo(client, pvc.GetName(), pvc.GetNamespace())
	if err != nil {
		log.Error(err, "unable to get runtimeInfo, get failure", "runtime", pvc.GetName(), "namespace", pvc.GetNamespace())
		return
	}
	runtimeInfo.SetDeprecatedNodeLabel(false)
	return
}
