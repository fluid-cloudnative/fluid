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
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	securityutil "github.com/fluid-cloudnative/fluid/pkg/utils/security"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
)

func (e *CacheEngine) BindToDataset() (err error) {
	e.Log.V(1).Info("Start to BindToDataset")

	return e.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}

func (e *CacheEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
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

		// TODO(cache runtime): update datasetToUpdate.Status.CacheStates

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

func (e *CacheEngine) generateDatasetMountOptions(m *datav1alpha1.Mount, sharedEncryptOptions []datav1alpha1.EncryptOption,
	sharedOptions map[string]string) (map[string]string, error) {

	// initialize mount options
	mOptions := map[string]string{}
	for k, v := range sharedOptions {
		mOptions[k] = v
	}

	for key, value := range m.Options {
		mOptions[key] = value
	}

	// if encryptOptions have the same key with options, it will overwrite the corresponding value
	var err error
	mOptions, err = e.genEncryptOptions(sharedEncryptOptions, mOptions, m.Name)
	if err != nil {
		return mOptions, err
	}

	// gen public encryptOptions
	mOptions, err = e.genEncryptOptions(m.EncryptOptions, mOptions, m.Name)
	if err != nil {
		return mOptions, err
	}

	return mOptions, nil
}

func (e *CacheEngine) genEncryptOptions(EncryptOptions []datav1alpha1.EncryptOption, mOptions map[string]string, name string) (map[string]string, error) {
	for _, item := range EncryptOptions {
		if _, ok := mOptions[item.Name]; ok {
			err := fmt.Errorf("the option %s is set more than one times, please double check the dataset's option and encryptOptions", item.Name)
			return mOptions, err
		}

		securityutil.UpdateSensitiveKey(item.Name)
		sRef := item.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(e.Client, sRef.Name, e.namespace)
		if err != nil {
			e.Log.Error(err, "get secret by mount encrypt options failed", "name", item.Name)
			return mOptions, err
		}

		e.Log.Info("get value from secret", "mount name", name, "secret key", sRef.Key)

		v := secret.Data[sRef.Key]
		mOptions[item.Name] = string(v)
	}

	return mOptions, nil
}
