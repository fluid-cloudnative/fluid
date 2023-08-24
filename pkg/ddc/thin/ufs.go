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
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
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

	// 2. update fuse-conf configmap
	err = t.updateFuseConfigOnChange(t.runtime, dataset)
	if err != nil {
		t.Log.Error(err, "Failed to update fuse config")
		return
	}
	return
}

func (t ThinEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	return true, nil
}

func (t ThinEngine) updateFuseConfigOnChange(runtime *datav1alpha1.ThinRuntime, dataset *datav1alpha1.Dataset) error {
	fuseConfigStorage := getFuseConfigStorage()
	if fuseConfigStorage != "configmap" {
		t.Log.Info("no need to update fuse config", "fuseConfigStorage", fuseConfigStorage)
		return nil
	}

	fuseConfigMapName := t.getFuseConfigMapName()
	fuseConfigMap, err := kubeclient.GetConfigmapByName(t.Client, fuseConfigMapName, t.namespace)
	if err != nil {
		return err
	}
	if fuseConfigMap == nil {
		// fuse configmap not found
		t.Log.Info("Fuse configmap not found", "fuseConfigMapName", fuseConfigMapName)
		return nil
	}

	configStr, err := t.transformFuseConfig(runtime, dataset, &ThinValue{})
	if err != nil {
		return err
	}

	fuseConfigMapToUpdate := fuseConfigMap.DeepCopy()
	fuseConfigMapToUpdate.Data["config.json"] = configStr
	if !reflect.DeepEqual(fuseConfigMap, fuseConfigMapToUpdate) {
		t.Log.Info("Update fuse config", "fuse config", configStr)
		err = kubeclient.UpdateConfigMap(t.Client, fuseConfigMapToUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t ThinEngine) getFuseConfigMapName() string {
	return t.name + "-fuse-conf"
}
