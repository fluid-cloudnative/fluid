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
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
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
	sharedOptions map[string]string) (mOptions map[string]string, encryptOptions map[string]string, err error) {

	// Initialize return maps
	mOptions = make(map[string]string)
	encryptOptions = make(map[string]string)

	// initialize mount options, mount options will overwrite shared options.
	for k, v := range sharedOptions {
		mOptions[k] = v
	}
	for key, value := range m.Options {
		mOptions[key] = value
	}

	// collect encrypt options, mount options will overwrite shared options.
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

		// Construct the secret mount path in the container
		// The secret will be mounted at /etc/fluid/secrets/<secret-name>/<key>
		secretPath := getSecretFilePath(sRef.Name, sRef.Key)

		// Store in map: key is option name, value is secret path
		existingEncryptOpts[item.Name] = secretPath
	}
	return nil
}
