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

package docker

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

const ImageTagEnvRegexFormat = "^\\S+:\\S+$"

var ImageTagEnvRegex = regexp.MustCompile(ImageTagEnvRegexFormat)

// ParseDockerImage extracts repo and tag from image. An empty string is returned if no tag is discovered.
func ParseDockerImage(image string) (name string, tag string) {
	matches := strings.Split(image, ":")
	if len(matches) >= 2 {
		name = strings.Replace(strings.Trim(fmt.Sprint(matches[0:len(matches)-1]), "[]"), " ", ":", -1)
		tag = matches[len(matches)-1]
	} else if len(matches) == 1 {
		name = matches[0]
		tag = "latest"
		// return matches[0], "latest"
	}
	return
}

// GetImageRepoFromEnv parse the image from environment variables, if it's not existed, return the default value
func GetImageRepoFromEnv(envName string) (image string) {
	if value, existed := os.LookupEnv(envName); existed {
		if matched := ImageTagEnvRegex.MatchString(value); matched {
			k, _ := ParseDockerImage(value)
			if len(k) > 0 {
				image = k
			}
		}
	}
	return
}

// GetImageTagFromEnv parse the image tag from environment variables, if it's not existed, return the default value
func GetImageTagFromEnv(envName string) (tag string) {
	if value, existed := os.LookupEnv(envName); existed {
		if matched := ImageTagEnvRegex.MatchString(value); matched {
			_, v := ParseDockerImage(value)
			if len(v) > 0 {
				tag = v
			}
		}
	}
	return
}

// get docker pull secrets from environment variables, if it's not existed, return []
// image pull secret format in ENV: str1,str2,str3
func GetImagePullSecretsFromEnv(envName string) []corev1.LocalObjectReference {
	imagePullSecrets := []corev1.LocalObjectReference{}
	if value, existed := os.LookupEnv(envName); existed {
		if len(value) > 0 {
			secrets := strings.Split(value, ",")
			for _, item := range secrets {
				if len(item) > 0 {
					imagePullSecrets = append(imagePullSecrets, corev1.LocalObjectReference{Name: item})
				}
			}
		}
	}
	return imagePullSecrets
}

// ParseInitImage parses the init image and image tag
func ParseInitImage(image, tag, imagePullPolicy, envName string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = GetImageRepoFromEnv(envName)
		if len(image) == 0 {
			initImageInfo := strings.Split(common.DefaultInitImage, ":")
			if len(initImageInfo) < 1 {
				panic("invalid default init image!")
			} else {
				image = initImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = GetImageTagFromEnv(envName)
		if len(tag) == 0 {
			initImageInfo := strings.Split(common.DefaultInitImage, ":")
			if len(initImageInfo) < 2 {
				panic("invalid default init image!")
			} else {
				tag = initImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

// GetWorkerImage get the image of alluxio worker from alluxioruntime, env or default
// TODO: Get image by calling runtime controller interface instead of reading runtime object
func GetWorkerImage(client client.Client, datasetName string, runtimeType string, namespace string) (imageName string, imageTag string) {
	configmapName := datasetName + "-" + runtimeType + "-values"
	configmap, err := kubeclient.GetConfigmapByName(client, configmapName, namespace)
	if configmap != nil && err == nil {
		for key, value := range configmap.Data {
			if key == "data" {
				splits := strings.Split(value, "\n")
				for _, split := range splits {
					if strings.HasPrefix(split, "image: ") {
						imageName = strings.TrimPrefix(split, "image: ")
					}
					if strings.HasPrefix(split, "imageTag: ") {
						imageTag = strings.TrimPrefix(split, "imageTag: ")
					}
				}
			}
		}
	}
	return
}
