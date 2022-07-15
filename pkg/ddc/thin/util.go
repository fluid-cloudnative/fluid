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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

// getRuntime gets thin runtime
func (t *ThinEngine) getRuntime() (*datav1alpha1.ThinRuntime, error) {

	key := types.NamespacedName{
		Name:      t.name,
		Namespace: t.namespace,
	}

	var runtime datav1alpha1.ThinRuntime
	if err := t.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (t *ThinEngine) getThinProfile() (profile *datav1alpha1.ThinProfile, err error) {
	if t.runtime == nil {
		return
	}
	key := types.NamespacedName{
		Name: t.runtime.Spec.ThinProfileName,
	}

	if err := t.Get(context.TODO(), key, profile); err != nil {
		return nil, err
	}
	return
}

func (t *ThinEngine) getFuseDaemonsetName() (dsName string) {
	return t.name + "-fuse"
}

func (t *ThinEngine) getWorkerName() (dsName string) {
	return t.name + "-worker"
}

func (t *ThinEngine) getMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	t.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/thin-fuse", mountRoot, t.namespace, t.name)
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.ThinRuntime
	} else {
		path = path + "/" + common.ThinRuntime
	}
	return
}
