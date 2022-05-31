package mutating

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/cache"
	corev1 "k8s.io/api/core/v1"
)

func CheckIfPVCIsDataset(pvc *corev1.PersistentVolumeClaim) (isDataset bool) {

	cache.GetCachedInfoForPersistentVolumeClaim(pvc)

	return kubeclient.CheckIfPVCIsDataset(pvc)
}
