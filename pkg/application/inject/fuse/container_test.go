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

package fuse

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/pod"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestFindInjectedSidecars_NoSidecars(t *testing.T) {
	pod1 := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test"},
				{Name: "test2"},
			},
		},
	}
	podObjs, err := pod.NewApplication(pod1).GetPodSpecs()
	assert.NoError(t, err)

	injectedSidecars, err := findInjectedSidecars(podObjs[0])
	assert.NoError(t, err)
	assert.Empty(t, injectedSidecars)
}

func TestFindInjectedSidecars_OneSidecar(t *testing.T) {
	pod2 := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "fluid-fuse-0"},
				{Name: "test"},
			},
		},
	}
	podObjs, err := pod.NewApplication(pod2).GetPodSpecs()
	assert.NoError(t, err)

	injectedSidecars, err := findInjectedSidecars(podObjs[0])
	assert.NoError(t, err)
	assert.Len(t, injectedSidecars, 1)
	assert.Equal(t, "fluid-fuse-0", injectedSidecars[0].Name)
}

func TestFindInjectedSidecars_MultipleSidecars(t *testing.T) {
	pod3 := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "fluid-fuse-0"},
				{Name: "test"},
				{Name: "fluid-fuse-1"},
				{Name: "fluid-fuse-dataset-xyz"},
			},
		},
	}
	podObjs, err := pod.NewApplication(pod3).GetPodSpecs()
	assert.NoError(t, err)

	injectedSidecars, err := findInjectedSidecars(podObjs[0])
	assert.NoError(t, err)
	assert.Len(t, injectedSidecars, 3)
	assert.Equal(t, "fluid-fuse-0", injectedSidecars[0].Name)
	assert.Equal(t, "fluid-fuse-1", injectedSidecars[1].Name)
	assert.Equal(t, "fluid-fuse-dataset-xyz", injectedSidecars[2].Name)
}

func TestFindInjectedSidecars_PrefixOnly(t *testing.T) {
	pod4 := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test-fluid-fuse"},
				{Name: "fluid-fuse-0"},
			},
		},
	}
	podObjs, err := pod.NewApplication(pod4).GetPodSpecs()
	assert.NoError(t, err)

	injectedSidecars, err := findInjectedSidecars(podObjs[0])
	assert.NoError(t, err)
	assert.Len(t, injectedSidecars, 1)
	assert.Equal(t, "fluid-fuse-0", injectedSidecars[0].Name)
}
