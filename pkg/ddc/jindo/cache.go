package jindo

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// queryCacheStatus checks the cache status
func (e *JindoEngine) queryCacheStatus() (states cacheStates, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Info("Failed to get dataset when query cache status")
		return states, err
	}

	// mock
	dataset.Status.UfsTotal = "100KiB"
	states.cacheCapacity = "10KiB"
	states.cached = "20KiB"

	// `dataset.Status.UfsTotal` probably haven't summed, in which case we won't compute cache percentage
	if dataset.Status.UfsTotal != "" && dataset.Status.UfsTotal != METADATA_SYNC_NOT_DONE_MSG {
		usedInBytes, _ := utils.FromHumanSize(states.cached)
		ufsTotalInBytes, _ := utils.FromHumanSize(dataset.Status.UfsTotal)
		states.cachedPercentage = fmt.Sprintf("%.1f%%", float64(usedInBytes)/float64(ufsTotalInBytes)*100.0)
	}

	return states, nil
}


// clean cache
func (e *JindoEngine) invokeCleanCache(path string) (err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUitls.CleanCache(path)
}
