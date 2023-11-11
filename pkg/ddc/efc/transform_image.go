/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package efc

import (
	corev1 "k8s.io/api/core/v1"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
)

func (e *EFCEngine) parseMasterImage(image string, tag string, imagePullPolicy string, imagePullSecrets []corev1.LocalObjectReference) (string, string, string, []corev1.LocalObjectReference) {
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

	if len(imagePullSecrets) == 0 {
		// if the environment variable is not set, it is still an empty slice
		imagePullSecrets = docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)
	}

	return image, tag, imagePullPolicy, imagePullSecrets
}

func (e *EFCEngine) parseWorkerImage(image string, tag string, imagePullPolicy string, imagePullSecrets []corev1.LocalObjectReference) (string, string, string, []corev1.LocalObjectReference) {
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

	if len(imagePullSecrets) == 0 {
		// if the environment variable is not set, it is still an empty slice
		imagePullSecrets = docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)
	}

	return image, tag, imagePullPolicy, imagePullSecrets
}

func (e *EFCEngine) parseFuseImage(image string, tag string, imagePullPolicy string, imagePullSecrets []corev1.LocalObjectReference) (string, string, string, []corev1.LocalObjectReference) {
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

	if len(imagePullSecrets) == 0 {
		// if the environment variable is not set, it is still an empty slice
		imagePullSecrets = docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)
	}

	return image, tag, imagePullPolicy, imagePullSecrets
}

func (e *EFCEngine) parseInitFuseImage(image string, tag string, imagePullPolicy string, imagePullSecrets []corev1.LocalObjectReference) (string, string, string, []corev1.LocalObjectReference) {
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

	if len(imagePullSecrets) == 0 {
		// if the environment variable is not set, it is still an empty slice
		imagePullSecrets = docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)
	}

	return image, tag, imagePullPolicy, imagePullSecrets
}
