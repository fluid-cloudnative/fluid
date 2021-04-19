package docker

import (
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
		if runtimeType == common.ALLUXIO_RUNTIME {
			imageName = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio"
		}
		if runtimeType == common.JINDO_RUNTIME {
			imageName = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata"
		}
	}
	if imageTag == "" {
		if runtimeType == common.ALLUXIO_RUNTIME {
			imageTag = "2.3.0-SNAPSHOT-238b7eb"
		}
		if runtimeType == common.JINDO_RUNTIME {
			imageTag = "3.5.0"
		}
	}
	return
}
