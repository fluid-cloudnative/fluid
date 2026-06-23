/*
  Copyright 2026 The Fluid Authors.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (e *CacheEngine) BindToDataset(runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) (err error) {
	e.Log.V(1).Info("Start to BindToDataset")

	err = e.updateMountTime()
	if err != nil {
		return
	}

	return e.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase, runtime, runtimeClass)
}

func (e *CacheEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()
		var cond datav1alpha1.DatasetCondition

		switch phase {
		case datav1alpha1.BoundDatasetPhase:
			if len(datasetToUpdate.Status.Mounts) == 0 {
				datasetToUpdate.Status.Mounts = datasetToUpdate.Spec.Mounts
			}

			if len(datasetToUpdate.Status.Runtimes) == 0 {
				datasetToUpdate.Status.Runtimes = []datav1alpha1.Runtime{}
			}

			datasetToUpdate.Status.Runtimes = utils.AddRuntimesIfNotExist(datasetToUpdate.Status.Runtimes, utils.NewRuntime(e.name,
				e.namespace,
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

		datasetToUpdate.Status.CacheStates = e.GetCacheStates(runtime, runtimeClass)

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			e.Log.Info("the dataset status", "status", datasetToUpdate.Status)
			err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return utils.LoggingErrorExceptConflict(e.Log, err, "Failed to Update dataset",
			types.NamespacedName{Namespace: e.namespace, Name: e.name})
	}

	return
}

func (e *CacheEngine) GetCacheStates(runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) common.CacheStateList {

	handler := resolveArchitectureApi(e.name, e.namespace, runtime, runtimeClass)

	executionEntries := handler.GetExecutionEntries()
	if executionEntries == nil || executionEntries.ReportSummary == nil {
		e.Log.Info("ReportSummary command is empty or not configured for this cache runtime",
			"name", e.name, "namespace", e.namespace)
		return nil
	}
	reportSummaryEntry := executionEntries.ReportSummary

	timeout := max(reportSummaryEntry.TimeoutSeconds, common.MinExecutionTimeoutSeconds)

	// Resolve target pod/container according to the runtime architecture.
	podName, containerName, err := handler.GetExecutionPodInfo()
	if err != nil {
		e.Log.Error(err, "Failed to get pod info")
		return nil
	}

	if podName == "" {
		e.Log.Info("No pod available, can not get cache states")
		return nil
	}

	cacheFileUtil := NewCacheFileUtil(podName, containerName, e.namespace, e.Log)
	stdout, err := cacheFileUtil.Mount(reportSummaryEntry.Command, time.Duration(timeout)*time.Second)
	if err != nil {
		e.Log.Error(err, "Failed to execute ReportSummary command", "stdout", stdout)
		return nil
	}

	var reportSummary datav1alpha1.CacheRuntimeReportSummary
	err = json.Unmarshal([]byte(stdout), &reportSummary)
	if err != nil {
		e.Log.Error(err, "Failed to unmarshal ReportSummary output", "stdout", stdout)
		return nil
	}

	cacheStates := make(common.CacheStateList)
	cacheStates[common.Cached] = reportSummary.Cached
	cacheStates[common.CachedPercentage] = reportSummary.CachedPercentage
	cacheStates[common.CacheCapacity] = reportSummary.CacheCapacity
	cacheStates[common.CacheHitRatio] = reportSummary.CacheHitRatio

	return cacheStates
}

func (e *CacheEngine) generateDatasetMountOptions(m *datav1alpha1.Mount, sharedEncryptOptions []datav1alpha1.EncryptOption,
	sharedOptions map[string]string) (mOptions map[string]string, encryptOptions map[string]string, err error) {

	mOptions = make(map[string]string)
	encryptOptions = make(map[string]string)

	for k, v := range sharedOptions {
		mOptions[k] = v
	}
	for key, value := range m.Options {
		mOptions[key] = value
	}

	err = e.collectEncryptOptions(sharedEncryptOptions, encryptOptions)
	if err != nil {
		return
	}
	err = e.collectEncryptOptions(m.EncryptOptions, encryptOptions)
	if err != nil {
		return
	}

	return
}

func (e *CacheEngine) collectEncryptOptions(encryptOpts []datav1alpha1.EncryptOption, existingEncryptOpts map[string]string) error {

	for _, item := range encryptOpts {
		sRef := item.ValueFrom.SecretKeyRef
		if sRef.Name == "" || sRef.Key == "" {
			return fmt.Errorf("encryptOption %s has empty secretKeyRef name or key", item.Name)
		}

		secretPath := getSecretFilePath(sRef.Name, sRef.Key)

		existingEncryptOpts[item.Name] = secretPath
	}
	return nil
}
