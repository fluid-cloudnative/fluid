/*
Copyright 2021 The Fluid Authors.

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

package unstructured

import (
	"github.com/mitchellh/mapstructure"
	corev1 "k8s.io/api/core/v1"
)

func extractToContainer(input map[string]interface{}) (container corev1.Container, err error) {
	container = corev1.Container{}
	err = mapstructure.Decode(input, &container)
	return
}

func extractToVolume(input map[string]interface{}) (volume corev1.Volume, err error) {
	volume = corev1.Volume{}
	err = mapstructure.Decode(input, &volume)
	return
}

func convertContainerToMap(container *corev1.Container) (output map[string]interface{}, err error) {
	output = map[string]interface{}{}
	if container != nil {
		err = mapstructure.Decode(container, &output)
	}
	return
}

func convertVolumeToMap(volume *corev1.Volume) (output map[string]interface{}, err error) {
	output = map[string]interface{}{}
	if volume != nil {
		err = mapstructure.Decode(volume, &output)
	}
	return
}
