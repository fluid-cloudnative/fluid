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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	"k8s.io/apimachinery/pkg/types"
)

// getRuntime get the current runtime
func (e *CacheEngine) getRuntime() (*datav1alpha1.CacheRuntime, error) {
	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.CacheRuntime
	if err := e.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}

	return &runtime, nil
}

func (e *CacheEngine) getRuntimeClass(runtimeClassName string) (*datav1alpha1.CacheRuntimeClass, error) {
	key := types.NamespacedName{
		Name: runtimeClassName,
	}
	var runtimeClass datav1alpha1.CacheRuntimeClass
	if err := e.Get(context.TODO(), key, &runtimeClass); err != nil {
		return nil, err
	}

	return &runtimeClass, nil
}

// getRuntimeInfo get the runtime info, may be called before dataset bound, so can not use base.GetRuntimeInfo but has
// the same processing logic.
func (e *CacheEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}
		opts := []base.RuntimeInfoOption{
			// TODO(cache runtime): useless code?
			base.WithTieredStore(datav1alpha1.TieredStore{}),
			// below used for create volume
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
			base.WithAnnotations(runtime.Annotations),
		}
		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, opts...)
		if err != nil {
			return e.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		e.runtimeInfo.SetFuseNodeSelector(runtime.Spec.Client.NodeSelector)
	}

	if testutil.IsUnitTest() {
		return e.runtimeInfo, nil
	}

	// Handling information of bound dataset. XXXEngine.getRuntimeInfo() might be called before the runtime is bound to a dataset,
	// so here we must lazily set dataset-related information once we found there's one bound dataset.
	if len(e.runtimeInfo.GetOwnerDatasetUID()) == 0 {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}

		uid, err := base.GetOwnerDatasetUIDFromRuntimeMeta(runtime.ObjectMeta)
		if err != nil {
			return nil, err
		}

		if len(uid) > 0 {
			e.runtimeInfo.SetOwnerDatasetUID(uid)
		}
	}

	if !e.runtimeInfo.IsPlacementModeSet() {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if utils.IgnoreNotFound(err) != nil {
			return nil, err
		}

		if dataset != nil {
			e.runtimeInfo.SetupWithDataset(dataset)
		}
	}

	return e.runtimeInfo, nil
}
