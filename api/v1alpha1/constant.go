/*
Copyright 2020 The Fluid Author.

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

package v1alpha1

const (
	// The cache system is ready
	DatasetReadyReason = "DatasetReady"

	// The cache system is updating
	DatasetUpdatingReason = "DatasetUpdating"

	// The cache system is failing
	DatasetDataSetFailedReason = "DatasetFailed"

	// The cache system fails to bind
	DatasetFailedToSetupReason = "DatasetFailedToSetup"
)

type PlacementMode string

const (
	ExclusiveMode PlacementMode = "Exclusive"

	ShareMode PlacementMode = "Shared"

	// DefaultMode is exclusive
	DefaultMode PlacementMode = ""
)

type FuseCleanPolicy string

const (
	// NoneCleanPolicy is the default clean policy. It will be transformed to OnRuntimeDeletedCleanPolicy automatically.
	NoneCleanPolicy FuseCleanPolicy = ""

	// OnDemandCleanPolicy cleans fuse pod once th fuse pod on some node is not needed
	OnDemandCleanPolicy FuseCleanPolicy = "OnDemand"

	// OnRuntimeDeletedCleanPolicy cleans fuse pod only when the cache runtime is deleted
	OnRuntimeDeletedCleanPolicy FuseCleanPolicy = "OnRuntimeDeleted"
)
