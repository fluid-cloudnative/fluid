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

func extractContainer(v map[string]interface{}) (container corev1.Container) {
	container = corev1.Container{}
	mapstructure.Decode(v, &container)
	return container
}

func extractVolumes(v map[string]interface{}) (volume corev1.Volume) {
	volume = corev1.Volume{}
	mapstructure.Decode(v, &volume)
	return volume
}
