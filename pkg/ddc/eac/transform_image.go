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

package eac

import (
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
)

func (e *EACEngine) parseMasterImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EACMasterImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACMasterImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default eac master image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EACMasterImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACMasterImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default eac master image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EACEngine) parseWorkerImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EACWorkerImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACWorkerImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default eac worker image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EACWorkerImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACWorkerImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default eac worker image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EACEngine) parseFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EACFuseImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACFuseImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default eac fuse image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EACFuseImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACFuseImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default eac fuse image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EACEngine) parseInitFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EACInitFuseImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACInitFuseImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default eac init alifuse image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EACInitFuseImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACInitFuseImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default eac init alifuse image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}
