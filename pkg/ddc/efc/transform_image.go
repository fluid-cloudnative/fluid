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

package efc

import (
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
)

func (e *EFCEngine) parseMasterImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EFCMasterImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEFCMasterImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default efc master image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EFCMasterImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEFCMasterImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default efc master image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EFCEngine) parseWorkerImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EFCWorkerImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEFCWorkerImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default efc worker image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EFCWorkerImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEFCWorkerImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default efc worker image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EFCEngine) parseFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EFCFuseImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEFCFuseImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default efc fuse image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EFCFuseImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEFCFuseImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default efc fuse image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EFCEngine) parseInitFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EFCInitFuseImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEFCInitFuseImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default efc init alifuse image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EFCInitFuseImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEFCInitFuseImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default efc init alifuse image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}
