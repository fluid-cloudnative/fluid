/*

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

package base

import (
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// SyncReplicas syncs the replicas
func (t *TemplateEngine) Sync(ctx cruntime.ReconcileRequestContext) (err error) {
	err = t.Implement.SyncMetadata()
	if err != nil {
		return
	}

	_, err = t.Implement.CheckAndUpdateRuntimeStatus()
	if err != nil {
		return
	}

	err = t.Implement.UpdateCacheOfDataset()
	if err != nil {
		return
	}

	// 1. Check healthy
	err = t.Implement.CheckRuntimeHealthy()
	if err != nil {
		return
	}

	// 2. Sync replicas
	err = t.Implement.SyncReplicas(ctx)
	if err != nil {
		return
	}

	// 3. Update runtime
	_, err = t.Implement.CheckAndUpdateRuntimeStatus()
	if err != nil {
		return
	}

	// 4. Update dataset
	_, err = t.Implement.UpdateOnUFSChange()
	if err != nil {
		return
	}

	return
}
