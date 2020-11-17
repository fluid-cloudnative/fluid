package docker

import (
	"os"
	"regexp"
	"strings"
)

// ParseDockerImage extracts repo and tag from image. An empty string is returned if no tag is discovered.
func ParseDockerImage(image string) (name string, tag string) {
	matches := strings.Split(image, ":")
	if len(matches) >= 2 {
		return matches[0], matches[1]
	} else if len(matches) == 1 {
		return matches[0], "latest"
	}
	return "", ""
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
