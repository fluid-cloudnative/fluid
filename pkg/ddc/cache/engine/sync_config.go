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

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	componenthelper "github.com/fluid-cloudnative/fluid/pkg/ddc/cache/componenthelper"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"reflect"
)

func (t *CacheEngine) syncCacheRuntimeConfig(dataset *datav1alpha1.Dataset, value *common.CacheRuntimeValue) error {
	mounts := []common.MountConfig{}
	for _, m := range dataset.Spec.Mounts {
		mountCig := common.MountConfig{
			Options:    utils.UnionMapsWithOverride(dataset.Spec.SharedOptions, m.Options),
			Name:       m.Name,
			MountPoint: m.MountPoint,
			ReadOnly:   m.ReadOnly,
			Shared:     m.Shared,
			Path:       m.Path,
		}

		mounts = append(mounts, mountCig)
	}

	config := common.CacheRuntimeConfig{
		Mounts:      mounts,
		AccessModes: dataset.Spec.AccessModes,
		Topology:    make(map[string]common.TopologyConfig),
	}
	if len(config.AccessModes) == 0 {
		config.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany}
	}

	if value.Master != nil && value.Master.Enabled {
		if t.masterHelper == nil {
			t.masterHelper = componenthelper.NewComponentHelper(value.Master.WorkloadType, t.Scheme, t.Client)
		}
		config.Master = &common.CacheRuntimeComponentConfig{
			Enabled: true,
			Options: value.Master.Options,
		}
		master, err := t.masterHelper.GetComponentTopologyInfo(context.TODO(), value.Master)
		if err != nil {
			return err
		}
		config.Topology["master"] = master
		if value.Master.EncryptOption != nil {
			config.Master.EncryptOption = value.Master.EncryptOption
		}

	}
	if value.Worker != nil && value.Worker.Enabled {
		if t.workerHelper == nil {
			t.workerHelper = componenthelper.NewComponentHelper(value.Worker.WorkloadType, t.Scheme, t.Client)
		}
		config.Worker = &common.CacheRuntimeComponentConfig{
			Enabled: true,
			Options: value.Worker.Options,
		}
		worker, err := t.workerHelper.GetComponentTopologyInfo(context.TODO(), value.Worker)
		if err != nil {
			return err
		}
		config.Topology["worker"] = worker
		if value.Worker.EncryptOption != nil {
			config.Worker.EncryptOption = value.Worker.EncryptOption
		}

		if value.Worker.TieredStore != nil {
			config.Worker.TieredStore = value.Worker.TieredStore
		}
	}
	if value.Client != nil && value.Client.Enabled {
		if t.clientHelper == nil {
			t.clientHelper = componenthelper.NewComponentHelper(value.Client.WorkloadType, t.Scheme, t.Client)
		}
		config.Client = &common.CacheRuntimeComponentConfig{
			Enabled:    true,
			TargetPath: t.getTargetPath(),
			Options:    value.Client.Options,
		}
		client, err := t.clientHelper.GetComponentTopologyInfo(context.TODO(), value.Client)
		if err != nil {
			return err
		}
		config.Topology["client"] = client
		if value.Client.EncryptOption != nil {
			config.Client.EncryptOption = value.Client.EncryptOption
		}

		if value.Client.TieredStore != nil {
			config.Client.TieredStore = value.Client.TieredStore
		}
	}
	b, _ := json.Marshal(config)
	configMap, err := kubeclient.GetConfigmapByName(t.Client, t.getRuntimeConfigCmName(), t.namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			configMapToCreate := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      t.getRuntimeConfigCmName(),
					Namespace: t.namespace,
				},
				Data: map[string]string{
					"config.json": string(b),
				},
			}
			return t.Client.Create(context.TODO(), configMapToCreate)
		}
		return err
	}
	if configMap == nil {
		return fmt.Errorf("fail to find ConfigMap name:%s, namespace:%s ", t.getRuntimeConfigCmName(), t.namespace)
	}
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		configMapToUpdate := configMap.DeepCopy()
		configMapToUpdate.Data["config.json"] = string(b)
		if !reflect.DeepEqual(configMapToUpdate, configMap) {
			err = t.Client.Update(context.TODO(), configMapToUpdate)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
