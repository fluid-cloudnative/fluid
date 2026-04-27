/*
Copyright 2020 The Kruise Authors.

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

package util

import (
	"strings"
)

// IsImageDigest indicates whether image is digest format,
// for example: docker.io/busybox@sha256:a9286defaba7b3a519d585ba0e37d0b2cbee74ebfe590960b0b1d6a5e97d1e1d
func IsImageDigest(image string) bool {
	return strings.Contains(image, "@")
}

// IsContainerImageEqual indicates whether container images are equal.
// Supports digest-format and tag-format images.
func IsContainerImageEqual(image1, image2 string) bool {
	if IsImageDigest(image1) && IsImageDigest(image2) {
		// Compare repo + digest
		repo1, digest1 := splitImageDigest(image1)
		repo2, digest2 := splitImageDigest(image2)
		return repo1 == repo2 && digest1 == digest2
	}
	if !IsImageDigest(image1) && !IsImageDigest(image2) {
		// Compare repo + tag
		repo1, tag1 := splitImageTag(image1)
		repo2, tag2 := splitImageTag(image2)
		return repo1 == repo2 && tag1 == tag2
	}
	return false
}

// splitImageDigest splits an image string into repo and digest parts.
func splitImageDigest(image string) (repo, digest string) {
	parts := strings.SplitN(image, "@", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return image, ""
}

// splitImageTag splits an image string into repo and tag parts.
func splitImageTag(image string) (repo, tag string) {
	// Remove digest part if present
	image = strings.SplitN(image, "@", 2)[0]
	// Find the last colon that is a tag separator
	// Handle cases like docker.io/library/ubuntu:20.04
	lastColon := strings.LastIndex(image, ":")
	lastSlash := strings.LastIndex(image, "/")
	if lastColon > lastSlash {
		return image[:lastColon], image[lastColon+1:]
	}
	return image, "latest"
}
