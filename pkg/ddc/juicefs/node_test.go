/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getTestJuiceFSEngineNode(client client.Client, name string, namespace string, withRunTime bool) *JuiceFSEngine {
	engine := &JuiceFSEngine{
		runtime:     nil,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: nil,
		Log:         fake.NullLogger(),
	}
	if withRunTime {
		engine.runtime = &datav1alpha1.JuiceFSRuntime{}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo(name, namespace, "alluxio", datav1alpha1.TieredStore{})
	}
	return engine
}

func TestSyncScheduleInfoToCacheNodes(t *testing.T) {
	engine := getTestJuiceFSEngineNode(nil, "test", "test", false)
	err := engine.SyncScheduleInfoToCacheNodes()
	if err != nil {
		t.Errorf("Failed with err %v", err)
	}
}
