package utils

import (
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/cache"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CollectRuntimeInfosFromPVCs(client client.Client, pvcNames []string, namespace string, setupLog logr.Logger, skipPrecheck bool) (runtimeInfos map[string]base.RuntimeInfoInterface, err error) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "CreateUpdatePodForSchedulingHandler.checkIfDatasetPVCs",
			"pvc.names", pvcNames, "pvc.namespace", namespace)
	}
	errPVCs := []string{}
	runtimeInfos = map[string]base.RuntimeInfoInterface{}
	for _, pvcName := range pvcNames {
		var (
			isDatasetPVC bool
			runtimeInfo  base.RuntimeInfoInterface
		)
		if cachedInfo, found := cache.GetRuntimeInfoByKey(types.NamespacedName{
			Name:      pvcName,
			Namespace: namespace,
		}); found {
			isDatasetPVC = cachedInfo.IsBelongToDataset()
			if isDatasetPVC {
				runtimeInfos[pvcName] = cachedInfo.GetRuntimeInfo()
			}
		} else {
			pvc, pvcErr := kubeclient.GetPersistentVolumeClaim(client, pvcName, namespace)
			if pvcErr != nil {
				setupLog.Error(pvcErr, "unable to check pvc, ignore and continue to check next pvc",
					"pvc",
					pvcName,
					"namespace",
					namespace)
				errPVCs = append(errPVCs, pvcName)
				continue
			}
			isDatasetPVC = kubeclient.CheckIfPVCIsDataset(pvc)
			if isDatasetPVC {
				runtimeInfo, err = buildRuntimeInfoInternalWithPrecheck(client, pvc, setupLog, skipPrecheck)
				if err != nil {
					err = errors.Wrapf(err, "failed to build runtime info for PVC \"%v/%v\"", namespace, pvcName)
					return
				}
				runtimeInfo.SetDeprecatedNodeLabel(false)
				runtimeInfos[pvcName] = runtimeInfo
			}
			cache.AddRuntimeInfoByKey(types.NamespacedName{
				Name:      pvcName,
				Namespace: namespace,
			}, runtimeInfo, isDatasetPVC)
		}
	}

	if len(errPVCs) > 0 {
		err = fmt.Errorf("failed to get the following PVCs %v", errPVCs)
		return
	}

	return
}

func buildRuntimeInfoInternalWithPrecheck(client client.Client,
	pvc *corev1.PersistentVolumeClaim,
	log logr.Logger, skipPrecheck bool) (runtimeInfo base.RuntimeInfoInterface, err error) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "mutating.buildRuntimeInfoInternalByPVC",
			"pvc.name", pvc.GetName(), "pvc.namespace", pvc.GetNamespace())
	}
	namespace := pvc.GetNamespace()
	if len(namespace) == 0 {
		namespace = corev1.NamespaceDefault
	}
	pvcName := pvc.GetName()
	if datasetName, exists := common.GetManagerDatasetFromLabels(pvc.Labels); exists {
		pvcName = datasetName
	}

	if !skipPrecheck {
		if err = checkDatasetBound(client, pvcName, namespace); err != nil {
			return
		}
	}

	runtimeInfo, err = base.GetRuntimeInfo(client, pvcName, namespace)
	if err != nil {
		log.Error(err, "unable to get runtimeInfo, get failure", "runtime", pvc.GetName(), "namespace", namespace)
		return
	}
	runtimeInfo.SetDeprecatedNodeLabel(false)
	return
}

func checkDatasetBound(client client.Client, name, namespace string) (err error) {
	dataset, err := utils.GetDataset(client, name, namespace)
	if err != nil {
		return
	}

	if dataset.Status.Phase == v1alpha1.NotBoundDatasetPhase || dataset.Status.Phase == v1alpha1.NoneDatasetPhase {
		_, cond := utils.GetDatasetCondition(dataset.Status.Conditions, v1alpha1.DatasetNotReady)
		if cond != nil {
			err = fmt.Errorf("dataset \"%s/%s\" not ready because %s", dataset.Namespace, dataset.Name, cond.Message)
			return
		}
		err = fmt.Errorf("dataset \"%s/%s\" not bound", dataset.Namespace, dataset.Name)
		return
	}
	return nil
}
