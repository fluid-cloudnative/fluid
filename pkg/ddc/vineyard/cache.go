/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// queryCacheStatus checks the cache status
func (e *VineyardEngine) queryCacheStatus(runtime *v1alpha1.VineyardRuntime) (states cacheStates, err error) {
	var cachesize uint64

	if len(e.runtime.Spec.TieredStore.Levels) != 0 {
		cachesize, err = strconv.ParseUint(strconv.FormatInt(e.runtime.Spec.TieredStore.Levels[0].Quota.Value(), 10), 10, 64)
		if err != nil {
			return
		}
	}

	if cachesize != 0 {
		states.cacheCapacity = utils.BytesSize(float64(cachesize * uint64(runtime.Status.WorkerNumberReady)))
	}

	summary, err := e.GetReportSummary()
	if err != nil {
		e.Log.Error(err, "Failed to get Vineyard summary when query cache status")
		return states, err
	}

	if len(summary) == 0 {
		return states, errors.New("Vineyard summary is empty")
	}

	states.cached = e.ParseReportSummary(summary)

	cachedInBytes, _ := utils.FromHumanSize(states.cached)
	cacheCapacityInBytes, _ := utils.FromHumanSize(states.cacheCapacity)

	if cacheCapacityInBytes == 0 {
		states.cachedPercentage = "0%"
	} else {
		states.cachedPercentage = fmt.Sprintf("%.1f%%", float64(cachedInBytes)/float64(cacheCapacityInBytes)*100.0)
	}

	return
}
