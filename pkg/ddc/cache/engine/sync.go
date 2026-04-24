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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"reflect"
	"time"
)

func (e *CacheEngine) Sync(ctx cruntime.ReconcileRequestContext) (err error) {
	// sync the runtime value configmap
	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}

	err = e.syncRuntimeValueConfigMap(ctx, runtime)
	if err != nil {
		return err
	}

	// TODO: implement other logic

	// handle ufs change

	// sync runtime status

	// handle runtime spec change

	// sync metadata

	// SyncScheduleInfoToCacheNodes

	return nil
}

func (e *CacheEngine) syncRuntimeValueConfigMap(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.CacheRuntime) error {
	configMap, err := kubeclient.GetConfigmapByNameWithContext(ctx, e.Client, e.getRuntimeConfigConfigMapName(), e.namespace)
	if err != nil {
		return err
	}
	data, err := e.generateRuntimeConfigData(runtime)
	if err != nil {
		return err
	}

	var True = true
	owner := []metav1.OwnerReference{
		{
			APIVersion:         runtime.APIVersion,
			Kind:               runtime.Kind,
			Name:               runtime.Name,
			UID:                runtime.UID,
			Controller:         &True,
			BlockOwnerDeletion: &True,
		},
	}

	if configMap == nil {
		return kubeclient.CreateConfigMapWithOwnerWithContext(ctx, e.Client, e.getRuntimeConfigConfigMapName(), e.namespace, data, owner)
	}

	configMapToUpdate := configMap.DeepCopy()
	configMapToUpdate.Data = data
	if !reflect.DeepEqual(configMapToUpdate, configMap) {
		err = kubeclient.UpdateConfigMapWithContext(ctx, e.Client, configMapToUpdate)
		if err != nil {
			return err
		}
	}

	return err
}
func getSyncRetryDuration() (d *time.Duration, err error) {
	if value, existed := os.LookupEnv(syncRetryDurationEnv); existed {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return d, err
		}
		d = &duration
	}
	return
}
