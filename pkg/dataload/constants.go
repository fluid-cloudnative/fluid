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

package dataload

type DataLoadPhase string

const (
	DataLoadPhaseNone     DataLoadPhase = ""
	DataLoadPhasePending  DataLoadPhase = "Pending"
	DataLoadPhaseLoading  DataLoadPhase = "Loading"
	DataLoadPhaseComplete DataLoadPhase = "Complete"
	DataLoadPhaseFailed   DataLoadPhase = "Failed"
	DataLoadPhaseFinished DataLoadPhase = "Finished"
)

// DataLoadConditionType is a valid value for DataloadCondition.Type
type DataLoadConditionType string

// These are valid conditions of a Dataload.
const (
	// DataloadComplete means the Dataload has completed its execution.
	DataLoadComplete DataLoadConditionType = "Complete"
	// DataloadFailed means the Dataload has failed its execution.
	DataLoadFailed DataLoadConditionType = "Failed"
)

const (
	DATALOAD_FINALIZER     = "fluid-dataload-controller-finalizer"
	DATALOAD_CHART         = "fluid-dataloader"
	DATALOAD_DEFAULT_TTL   int64 = 60
	DATALOAD_DEFAULT_IMAGE = "registry.cn-hangzhou.aliyuncs.com/fluid/fluid-dataloader"
	DATALOAD_SUFFIX_LENGTH = 5
	ENV_DATALOADER_IMG     = "DATALOADER_IMG"
)
