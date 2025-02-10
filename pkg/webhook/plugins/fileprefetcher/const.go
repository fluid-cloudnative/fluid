/*
Copyright 2025 The Fluid Authors.

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

package fileprefetcher

// Environment variables for file prefetcher
const (
	envKeyFilePrefetcherFileList       = "FILE_PREFETCHER_FILE_LIST"
	envKeyFilePrefetcherAsyncPrefetch  = "FILE_PREFETCHER_ASYNC_PREFETCH"
	envKeyFilePrefetcherTimeoutSeconds = "FILE_PREFETCHER_TIMEOUT_SECONDS"

	envKeyFilePrefetcherImage = "FILE_PREFETCHER_IMAGE"
)

// Constants for file prefetcher
const (
	filePrefetcherContainerName         = "fluid-file-prefetcher"
	filePrefetcherStatusVolumeName      = "fluid-file-prefetcher-status-vol"
	filePrefetcherStatusVolumeMountPath = "/tmp/fluid-file-prefetcher/status"

	filePrefetcherDefaultFileList          = "<ALL>"
	filePrefetcherDefaultTimeoutSecondsStr = "120"
)
