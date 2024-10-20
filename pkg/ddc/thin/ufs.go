/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (t *ThinEngine) UsedStorageBytes() (int64, error) {
	return t.usedSpaceInternal()
}

func (t *ThinEngine) FreeStorageBytes() (int64, error) {
	return 0, nil
}

func (t *ThinEngine) TotalStorageBytes() (int64, error) {
	return t.totalStorageBytesInternal()
}

func (t *ThinEngine) TotalFileNums() (int64, error) {
	return t.totalFileNumsInternal()
}

func (t ThinEngine) ShouldCheckUFS() (should bool, err error) {
	return false, nil
}

func (t ThinEngine) PrepareUFS() (err error) {
	return
}

func (t ThinEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	// 1. get the dataset
	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		t.Log.Error(err, "Failed to get the dataset")
		return
	}
	if dataset == nil {
		t.Log.Info("Dataset not found", "Dataset name", t.name, "Dataset namespace", t.namespace)
		return
	}

	// 2. update fuse-conf configmap
	update, err := t.updateFuseConfigOnChange(t.runtime, dataset)
	if err != nil {
		t.Log.Error(err, "Failed to update fuse config")
		return
	}

	// 3. update fuse pod to sync configmap
	if update {
		err := t.updateFusePod()
		if err != nil {
			t.Log.Error(err, "Failed to update fuse pod")
			return
		}
	}
	return
}

func (t ThinEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	return true, nil
}

func (t ThinEngine) updateFuseConfigOnChange(runtime *datav1alpha1.ThinRuntime, dataset *datav1alpha1.Dataset) (update bool, err error) {
	fuseConfigStorage := getFuseConfigStorage()
	if fuseConfigStorage != "configmap" {
		t.Log.Info("no need to update fuse config", "fuseConfigStorage", fuseConfigStorage)
		return update, nil
	}

	fuseConfigMapName := t.getFuseConfigMapName()
	fuseConfigMap, err := kubeclient.GetConfigmapByName(t.Client, fuseConfigMapName, t.namespace)
	if err != nil {
		return update, err
	}
	if fuseConfigMap == nil {
		// fuse configmap not found
		t.Log.Info("Fuse configmap not found", "fuseConfigMapName", fuseConfigMapName)
		return update, nil
	}

	updatedThinValue := &ThinValue{}
	err = t.transformFuseConfig(runtime, dataset, updatedThinValue)
	if err != nil {
		return update, nil
	}

	fuseConfigMapToUpdate := fuseConfigMap.DeepCopy()
	fuseConfigMapToUpdate.Data["config.json"] = updatedThinValue.Fuse.ConfigValue
	if !reflect.DeepEqual(fuseConfigMap, fuseConfigMapToUpdate) {
		t.Log.Info("Update fuse config", "fuse config", updatedThinValue.Fuse.ConfigValue)
		err = kubeclient.UpdateConfigMap(t.Client, fuseConfigMapToUpdate)
		if err != nil {
			return update, err
		}
		update = true
	}
	return update, nil
}

// Updating the fuse pod is to let kubelet run syncPod and update the configmap content
func (t ThinEngine) updateFusePod() (err error) {
	// get fuse pod
	pods, err := t.GetRunningPodsOfDaemonset(t.getFuseDaemonsetName(), t.namespace)
	if err != nil {
		t.Log.Error(err, "Failed to get fuse pods")
		return
	}
	if len(pods) == 0 {
		return nil
	}
	// update pod annotation
	annotation := common.LabelAnnotationFusePrefix + "update-fuse-config"
	for _, pod := range pods {
		podToUpdate := pod.DeepCopy()
		annotations := podToUpdate.GetAnnotations()
		if val, ok := annotations[annotation]; ok {
			// if annotation has existed, add val
			count, err := strconv.Atoi(val)
			if err != nil {
				t.Log.Error(err, "Fuse pod has invalid annotation")
				return err
			}
			annotations[annotation] = strconv.Itoa(count + 1)
		} else {
			// if annotation not exist ,set val to 1
			annotations[annotation] = "1"
		}
		podToUpdate.SetAnnotations(annotations)

		err := t.Client.Update(context.TODO(), podToUpdate)
		if err != nil {
			t.Log.Error(err, fmt.Sprintf("Fuse pod %s update failed", podToUpdate.GetName()))
			return err
		}
	}
	return
}
