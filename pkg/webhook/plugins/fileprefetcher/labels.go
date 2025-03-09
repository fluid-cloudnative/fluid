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

import "github.com/fluid-cloudnative/fluid/pkg/common"

const (
	// prefix for labels and annotations used in file prefetcher feature
	// i.e. file-prefetcher.fluid.io/
	LabelAnnotationFilePrefetcherPrefix = "file-prefetcher." + common.LabelAnnotationPrefix

	// annotation to inject file prefetcher sidecar container
	// i.e. file-prefetcher.fluid.io/inject
	AnnotationFilePrefetcherInject = LabelAnnotationFilePrefetcherPrefix + "inject"

	// annotation to mark file prefetcher sidecar container has been injected
	// i.e. file-prefetcher.fluid.io/inject-done
	AnnotationFilePrefetcherInjectDone = LabelAnnotationFilePrefetcherPrefix + "inject-done"

	// annotation to set file prefetcher sidecar container's image
	// i.e. file-prefetcher.fluid.io/image
	AnnotationFilePrefetcherImage = LabelAnnotationFilePrefetcherPrefix + "image"

	// annotation to set file list to prefetch
	// i.e. file-prefetcher.fluid.io/file-list
	AnnotationFilePrefetcherFileList = LabelAnnotationFilePrefetcherPrefix + "file-list"

	// annotation to set if prefetch files asynchronously. When setting it to true,
	// app container will start up after file prefetcher finishes.
	// i.e. file-prefetcher.fluid.io/async-prefetch
	AnnotationFilePrefetcherAsync = LabelAnnotationFilePrefetcherPrefix + "async-prefetch"

	// annotation to set extra envs for file prefetcher
	// i.e. file-prefetcher.fluid.io/extra-envs
	AnnotationFilePrefetcherExtraEnvs = LabelAnnotationFilePrefetcherPrefix + "extra-envs"

	// annotation to set timeout for file prefetcher
	// i.e. file-prefetcher.fluid.io/prefetch-timeout-seconds
	AnnotationFilePrefetcherTimeoutSeconds = LabelAnnotationFilePrefetcherPrefix + "prefetch-timeout-seconds"
)
