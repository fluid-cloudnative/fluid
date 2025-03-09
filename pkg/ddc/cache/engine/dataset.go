package engine

import (
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
)

func (t *CacheEngine) BindToDataset() (err error) {
	t.Log.V(1).Info("Start to BindToDataset")

	return t.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}

func (t *CacheEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()
		var cond datav1alpha1.DatasetCondition

		switch phase {
		case datav1alpha1.BoundDatasetPhase:
			// Stores dataset mount info
			if len(datasetToUpdate.Status.Mounts) == 0 {
				datasetToUpdate.Status.Mounts = datasetToUpdate.Spec.Mounts
			}

			// Stores binding relation between dataset and runtime
			if len(datasetToUpdate.Status.Runtimes) == 0 {
				datasetToUpdate.Status.Runtimes = []datav1alpha1.Runtime{}
			}

			datasetToUpdate.Status.Runtimes = utils.AddRuntimesIfNotExist(datasetToUpdate.Status.Runtimes, utils.NewRuntime(t.name,
				t.namespace,
				common.AccelerateCategory,
				common.CacheRuntime,
				0))

			cond = utils.NewDatasetCondition(datav1alpha1.DatasetReady, datav1alpha1.DatasetReadyReason,
				"The ddc runtime is ready.",
				corev1.ConditionTrue)
		case datav1alpha1.FailedDatasetPhase:
			cond = utils.NewDatasetCondition(datav1alpha1.DatasetReady, datav1alpha1.DatasetReadyReason,
				"The ddc runtime is not ready.",
				corev1.ConditionFalse)
		default:
			cond = utils.NewDatasetCondition(datav1alpha1.DatasetReady, datav1alpha1.DatasetReadyReason,
				"The ddc runtime is unknown.",
				corev1.ConditionFalse)
		}

		datasetToUpdate.Status.Phase = phase
		datasetToUpdate.Status.Conditions = utils.UpdateDatasetCondition(datasetToUpdate.Status.Conditions,
			cond)

		//datasetToUpdate.Status.CacheStates = runtime.Status.CacheStates

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = t.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				return err
			} else {
				t.Log.Info("No need to update the cache of the data")
			}
		}

		return nil
	})

	if err != nil {
		return utils.LoggingErrorExceptConflict(t.Log, err, "Failed to Update dataset",
			types.NamespacedName{Namespace: t.namespace, Name: t.name})
	}

	return
}

func (t *CacheEngine) UpdateCacheOfDataset() (err error) {
	return nil
}
