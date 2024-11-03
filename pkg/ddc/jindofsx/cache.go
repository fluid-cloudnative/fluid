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

package jindofsx

import (
	"fmt"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindofsx/operations"
	ddctypes "github.com/fluid-cloudnative/fluid/pkg/ddc/types"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// queryCacheStatus checks the cache status
func (e *JindoFSxEngine) queryCacheStatus() (states cacheStates, err error) {
	defer utils.TimeTrack(time.Now(), "JindoFSxEngine.queryCacheStatus", "name", e.name, "namespace", e.namespace)
	summary, err := e.GetReportSummary()
	if err != nil {
		e.Log.Error(err, "Failed to get Jindo summary when query cache status")
		return states, err
	}
	totalCapacityLabel := ""
	usedCapacityLabel := ""
	if len(e.runtime.Spec.TieredStore.Levels) > 0 && e.runtime.Spec.TieredStore.Levels[0].MediumType == "MEM" {
		totalCapacityLabel = SUMMARY_PREFIX_TOTAL_MEM_CAPACITY
		usedCapacityLabel = SUMMARY_PREFIX_USED_MEM_CAPACITY
	} else {
		totalCapacityLabel = ddctypes.SummaryPrefixTotalDiskCapacity
		usedCapacityLabel = ddctypes.SummaryPrefixUsedDiskCapacity
	}
	strs := strings.Split(summary, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, totalCapacityLabel) {
			totalCacheCapacityJindo, _ := utils.FromHumanSize(strings.TrimPrefix(str, totalCapacityLabel))
			// Convert JindoFS's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cacheCapacity = utils.BytesSize(float64(totalCacheCapacityJindo))
		}
		if strings.HasPrefix(str, usedCapacityLabel) {
			usedCacheCapacityJindo, _ := utils.FromHumanSize(strings.TrimPrefix(str, usedCapacityLabel))
			// Convert JindoFS's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cached = utils.BytesSize(float64(usedCacheCapacityJindo))
		}
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Info("Failed to get dataset when query cache status")
		return states, err
	}

	// `dataset.Status.UfsTotal` probably haven't summed, in which case we won't compute cache percentage
	if dataset.Status.UfsTotal != "" && dataset.Status.UfsTotal != METADATA_SYNC_NOT_DONE_MSG {
		usedInBytes, _ := utils.FromHumanSize(states.cached)
		ufsTotalInBytes, _ := utils.FromHumanSize(dataset.Status.UfsTotal)
		// jindofs calculate cached storage bytesize with block sum, so percentage will be over 100% if totally cached
		percentTage := 0.0
		if ufsTotalInBytes != 0 {
			percentTage = float64(usedInBytes) / float64(ufsTotalInBytes)
		}
		// avoid jindo blocksize greater than ufssize
		if percentTage > 1 {
			percentTage = 1.0
		}
		states.cachedPercentage = fmt.Sprintf("%.1f%%", percentTage*100.0)
	}

	return states, nil
}

// clean cache
func (e *JindoFSxEngine) invokeCleanCache() (err error) {
	// 1. Check if master is ready, if not, just return
	masterName := e.getMasterName()
	master, err := kubeclient.GetStatefulSet(e.Client, masterName, e.namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			e.Log.Info("Failed to get master", "err", err.Error())
			return nil
		}
		// other error
		return err
	}
	if master.Status.ReadyReplicas == 0 {
		e.Log.Info("The master is not ready, just skip clean cache.", "master", masterName)
		return nil
	} else {
		e.Log.Info("The master is ready, so start cleaning cache", "master", masterName)
	}

	// 2. run clean action
	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	e.Log.Info("cleaning cache and wait for a while")
	return fileUitls.CleanCache()
}
