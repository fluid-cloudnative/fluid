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

package common

type DataloadPhase string

const (
	DataloadPhaseNone     DataloadPhase = ""
	DataloadPhasePending  DataloadPhase = "Pending"
	DataloadPhaseLoading  DataloadPhase = "Loading"
	DataloadPhaseComplete DataloadPhase = "Complete"
	DataloadPhaseFailed   DataloadPhase = "Failed"
)

// DataloadConditionType is a valid value for DataloadCondition.Type
type DataloadConditionType string

// These are valid conditions of a Dataload.
const (
	// DataloadComplete means the Dataload has completed its execution.
	DataloadComplete DataloadConditionType = "Complete"
	// DataloadFailed means the Dataload has failed its execution.
	DataloadFailed DataloadConditionType = "Failed"
)

const (
	DATALOAD_FINALIZER     = "fluid-dataload-controller-finalizer"
	DATALOAD_CHART         = "fluid-dataloader"
	DATALOAD_DEFAULT_IMAGE = "registry.cn-hangzhou.aliyuncs.com/fluid/fluid-dataloader"
	DATALOAD_SUFFIX_LENGTH = 5
	ENV_DATALOADER_IMG     = "DATALOADER_IMG"
)
