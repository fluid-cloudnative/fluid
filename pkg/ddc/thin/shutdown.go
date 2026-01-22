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
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Shutdown performs a graceful shutdown of the ThinEngine with retry capabilities.
// The shutdown process includes cache cleanup, worker destruction, master destruction,
// and final resource cleanup. If cache cleanup fails, it will retry up to the
// gracefulShutdownLimits. Returns an error if any step in the shutdown sequence fails.
// The method executes steps sequentially and stops immediately on any error.
func (t ThinEngine) Shutdown() (err error) {
	if t.retryShutdown < t.gracefulShutdownLimits {
		err = t.cleanupCache()
		if err != nil {
			t.retryShutdown = t.retryShutdown + 1
			t.Log.Info("clean cache failed",
				"retry times", t.retryShutdown)
			return
		}
	}

	_, err = t.destroyWorkers(-1)
	if err != nil {
		return
	}

	err = t.destroyMaster()
	if err != nil {
		return
	}

	err = t.cleanAll()
	return
}

// destroyMaster Destroy the master
func (t *ThinEngine) destroyMaster() (err error) {
	var found bool
	found, err = helm.CheckRelease(t.name, t.namespace)
	if err != nil {
		return err
	}

	if found {
		err = helm.DeleteRelease(t.name, t.namespace)
		if err != nil {
			return
		}
	} else {
		// When upgrade Fluid to v1.0.0+ from a lower version, there may be some orphaned configmaps when deleting a ThinRuntime if it's created before the upgradation.
		// Detect such orphaned configmaps and clean them up.
		err = t.cleanUpOrphanedResources()
		if err != nil {
			t.Log.Info("WARNING: failed to delete orphaned resource, some resources may not be cleaned up in the cluster", "err", err)
			err = nil
		}
	}

	return
}

// cleanupCache cleans up the cache
func (t *ThinEngine) cleanupCache() (err error) {
	// todo
	return
}

// destroyWorkers attempts to delete the workers until worker num reaches the given expectedWorkers, if expectedWorkers is -1, it means all the workers should be deleted
// This func returns currentWorkers representing how many workers are left after this process.
func (t *ThinEngine) destroyWorkers(expectedWorkers int32) (currentWorkers int32, err error) {
	//  SchedulerMutex only for patch mode
	lifecycle.SchedulerMutex.Lock()
	defer lifecycle.SchedulerMutex.Unlock()

	runtimeInfo, err := t.getRuntimeInfo()
	if err != nil {
		return currentWorkers, err
	}

	return t.Helper.TearDownWorkers(runtimeInfo)
}

func (t *ThinEngine) cleanAll() (err error) {
	count, err := t.Helper.CleanUpFuse()
	if err != nil {
		t.Log.Error(err, "Err in cleaning fuse")
		return err
	}
	t.Log.Info("clean up fuse count", "n", count)

	var (
		valueConfigmapName = t.getHelmValuesConfigMapName()
		thinConfigmapName  = t.name + "-config"
		namespace          = t.namespace
	)

	cmNames := []string{valueConfigmapName, thinConfigmapName}

	for _, cmName := range cmNames {
		_, err = kubeclient.GetConfigmapByName(t.Client, cmName, namespace)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		err = kubeclient.DeleteConfigMap(t.Client, cmName, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

// runtimeset config has been duplicated, keep it for garbage collection
func (t *ThinEngine) cleanUpOrphanedResources() (err error) {
	orphanedConfigMapName := fmt.Sprintf("%s-runtimeset", t.name)
	cm, err := kubeclient.GetConfigmapByName(t.Client, orphanedConfigMapName, t.namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			t.Log.Info("Orphaned configmap not exist, do not need to delete it", "configmap", orphanedConfigMapName)
			return nil
		}
		return err
	}

	if cm != nil {
		if err = kubeclient.DeleteConfigMap(t.Client, orphanedConfigMapName, t.namespace); err != nil && utils.IgnoreNotFound(err) != nil {
			return err
		}
		t.Log.Info("Found orphaned configmap, successfully deleted it", "configmap", orphanedConfigMapName)
	}

	return nil
}
