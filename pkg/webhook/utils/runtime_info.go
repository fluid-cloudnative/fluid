package utils

import (
	"fmt"
	"strings"
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
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CollectRuntimeInfosFromPVCs(client client.Reader, pvcNames []string, namespace string, setupLog logr.Logger, skipPrecheck bool) (runtimeInfos map[string]base.RuntimeInfoInterface, err error) {
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

// CollectRuntimeInfosFromAnnotations collects runtime infos for the datasets a pod declares in its
// fluid.io/datasets annotation. It is the counterpart of CollectRuntimeInfosFromPVCs for pods that
// mount no dataset volume at all, e.g. app pods of a cache runtime that has no fuse client.
//
// The returned map is keyed by dataset name. Datasets are resolved in the pod's own namespace,
// because a pod can only reference volumes and config maps from there, so the annotation carries
// bare names rather than namespaced ones.
func CollectRuntimeInfosFromAnnotations(client client.Reader, annotations map[string]string, namespace string, setupLog logr.Logger, skipPrecheck bool) (runtimeInfos map[string]base.RuntimeInfoInterface, err error) {
	runtimeInfos = map[string]base.RuntimeInfoInterface{}

	datasetNames := annotations[common.LabelAnnotationDatasets]
	if len(datasetNames) == 0 {
		return
	}

	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "mutating.CollectRuntimeInfosFromAnnotations",
			"dataset.names", datasetNames, "dataset.namespace", namespace)
	}

	seen := map[string]struct{}{}
	errDatasets := []string{}
	for _, datasetName := range strings.Split(datasetNames, ",") {
		datasetName = strings.TrimSpace(datasetName)
		if len(datasetName) == 0 {
			continue
		}
		if _, dup := seen[datasetName]; dup {
			continue
		}
		seen[datasetName] = struct{}{}

		// A dataset name is a k8s object name, so an invalid one can never resolve. Reject it here
		// with a message naming the annotation, instead of letting it fail later as a not-found.
		if nameErrs := validation.IsDNS1035Label(datasetName); len(nameErrs) > 0 {
			setupLog.Error(errors.New(strings.Join(nameErrs, "; ")),
				"invalid dataset name in annotation, ignore and continue to check next dataset",
				"annotation", common.LabelAnnotationDatasets,
				"dataset", datasetName)
			errDatasets = append(errDatasets, datasetName)
			continue
		}

		if !skipPrecheck {
			if boundErr := checkDatasetBound(client, datasetName, namespace); boundErr != nil {
				setupLog.Error(boundErr, "unable to check dataset, ignore and continue to check next dataset",
					"dataset", datasetName,
					"namespace", namespace)
				errDatasets = append(errDatasets, datasetName)
				continue
			}
		}

		runtimeInfo, infoErr := base.GetRuntimeInfo(client, datasetName, namespace)
		if infoErr != nil {
			setupLog.Error(infoErr, "unable to get runtimeInfo, ignore and continue to check next dataset",
				"dataset", datasetName,
				"namespace", namespace)
			errDatasets = append(errDatasets, datasetName)
			continue
		}

		runtimeInfos[datasetName] = runtimeInfo
	}

	if len(errDatasets) > 0 {
		err = fmt.Errorf("failed to get the following datasets %v declared in annotation %s",
			errDatasets, common.LabelAnnotationDatasets)
		return
	}

	return
}

func buildRuntimeInfoInternalWithPrecheck(client client.Reader,
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
	return
}

func checkDatasetBound(client client.Reader, name, namespace string) (err error) {
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
