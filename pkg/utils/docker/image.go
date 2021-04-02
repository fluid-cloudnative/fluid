package docker

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"os"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// ParseDockerImage extracts repo and tag from image. An empty string is returned if no tag is discovered.
func ParseDockerImage(image string) (name string, tag string) {
	matches := strings.Split(image, ":")
	if len(matches) >= 2 {
		name = matches[0]
		tag = matches[1]
	} else if len(matches) == 1 {
		name = matches[0]
		tag = "latest"
		// return matches[0], "latest"
	}
	return
}

// GetImageRepoTagFromEnv parse the image and tag from environment varaibles, if it's not existed or
func GetImageRepoTagFromEnv(envName, defaultImage string, defaultTag string) (image, tag string) {

	image = defaultImage
	tag = defaultTag

	if value, existed := os.LookupEnv(envName); existed {
		if matched, err := regexp.MatchString("^\\S+:\\S+$", value); err == nil && matched {
			k, v := ParseDockerImage(value)
			if len(k) > 0 {
				image = k

			}
			if len(v) > 0 {
				tag = v

			}
		}
	}

	return
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
	if imageName == "" {
		imageName = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio"
	}
	if imageTag == "" {
		imageTag = "2.3.0-SNAPSHOT-238b7eb"
	}
	return
}

func GetInitUserImage(specImage common.ImageInfo) (Image string, ImageTag string, ImagePullPolicy string) {
	var initImage = ""
	if value, existed := os.LookupEnv(common.ALLUXIO_INIT_IMAGE_ENV); existed {
		if matched, err := regexp.MatchString("^\\S+:\\S+$", value); err == nil && matched {
			initImage = value
		}
	}
	if len(initImage) == 0 {
		initImage = common.DEFAULT_ALLUXIO_INIT_IMAGE
	}
	initImageInfo := strings.Split(initImage, ":")
	Image = initImageInfo[0]
	ImageTag = initImageInfo[1]
	ImagePullPolicy = "IfNotPresent"
	if len(specImage.Image) > 0 {
		Image = specImage.Image
	}

	if len(specImage.ImageTag) > 0 {
		ImageTag = specImage.ImageTag
	}

	if len(specImage.ImagePullPolicy) > 0 {
		ImagePullPolicy = specImage.ImagePullPolicy
	}
	return
}
