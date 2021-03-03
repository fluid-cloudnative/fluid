package jindo

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"strings"
)

// queryCacheStatus checks the cache status
func (e *JindoEngine) queryCacheStatus() (states cacheStates, err error) {
	summary, err := e.reportSummary()
	if err != nil {
		e.Log.Error(err, "Failed to get Jindo summary when query cache status")
		return states, err
	}
	strs := strings.Split(summary, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, SUMMARY_PREFIX_TOTAL_CAPACITY) {
			totalCacheCapacityJindo, _ := utils.FromHumanSize(strings.TrimPrefix(str, SUMMARY_PREFIX_TOTAL_CAPACITY))
			// Convert JindoFS's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cacheCapacity = utils.BytesSize(float64(totalCacheCapacityJindo))
		}
		if strings.HasPrefix(str, SUMMARY_PREFIX_USED_CAPACITY) {
			usedCacheCapacityJindo, _ := utils.FromHumanSize(strings.TrimPrefix(str, SUMMARY_PREFIX_USED_CAPACITY))
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
		// jindofs calculate cached storage bytesize with block sum, so precentage will be over 100% if totally cached
		percentTage := float64(usedInBytes) / float64(ufsTotalInBytes)
		if percentTage > 1 {
			percentTage = 1
		}
		states.cachedPercentage = fmt.Sprintf("%.1f%%", percentTage*100.0)
	}

	return states, nil
}
