package docker

import (
	"context"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
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

// GetAlluxioWorkerImage get the image of alluxio worker from alluxioruntime, env or default
// TODO: Get image by calling runtime controller interface instead of reading runtime object
func GetAlluxioWorkerImage(client client.Client, datasetName string, namespace string)(imageName string, imageTag string){
	alluxioruntime := &v1alpha1.AlluxioRuntime{}
	key := types.NamespacedName{
		Name: datasetName,
		Namespace: namespace,
	}
	err := client.Get(context.TODO(), key, alluxioruntime)
	if err != nil || alluxioruntime.Spec.AlluxioVersion.Image == "" || alluxioruntime.Spec.AlluxioVersion.ImageTag == "" {
		// when user have not define image url in env, will use the default
		imageName = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio"
		imageTag = "2.3.0-SNAPSHOT-238b7eb"
		imageName, imageTag = GetImageRepoTagFromEnv(common.ALLUXIO_RUNTIME_IMAGE_ENV, imageName, imageTag)
		return

	}
	// when user have define image url in alluxioruntime
	imageName = alluxioruntime.Spec.AlluxioVersion.Image
	imageTag = alluxioruntime.Spec.AlluxioVersion.ImageTag
	return

}
