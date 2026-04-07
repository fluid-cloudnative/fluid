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
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func newTestInjector(t *testing.T) *Injector {
	t.Helper()
	return &Injector{log: zap.New(zap.UseDevMode(true))}
}

// ---- appendVolumes tests ----

func TestAppendVolumes_NoConflict(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "volume1"}, {Name: "volume2"}}
	volumesToAdd := []corev1.Volume{{Name: "volume3"}, {Name: "volume4"}}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-0")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 4)
	assert.Len(t, conflictMap, 2)
	assert.Equal(t, "volume3-0", conflictMap["volume3"])
	assert.Equal(t, "volume4-0", conflictMap["volume4"])
	assert.Equal(t, "volume3-0", resultVolumes[2].Name)
	assert.Equal(t, "volume4-0", resultVolumes[3].Name)
}

func TestAppendVolumes_EmptyVolumesToAdd(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "volume1"}, {Name: "volume2"}}
	volumesToAdd := []corev1.Volume{}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-0")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 2)
	assert.Empty(t, conflictMap)
}

func TestAppendVolumes_EmptySuffix(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "volume1"}}
	volumesToAdd := []corev1.Volume{{Name: "volume2"}}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 2)
	assert.Empty(t, conflictMap)
	assert.Equal(t, "volume2", resultVolumes[1].Name)
}

func TestAppendVolumes_EmptyExistingVolumes(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{}
	volumesToAdd := []corev1.Volume{{Name: "volume1"}, {Name: "volume2"}}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-0")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 2)
	assert.Len(t, conflictMap, 2)
	assert.Equal(t, "volume1-0", conflictMap["volume1"])
	assert.Equal(t, "volume2-0", conflictMap["volume2"])
	assert.Equal(t, "volume1-0", resultVolumes[0].Name)
	assert.Equal(t, "volume2-0", resultVolumes[1].Name)
}

func TestAppendVolumes_SingleConflict(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "volume1"}, {Name: "volume2-0"}}
	volumesToAdd := []corev1.Volume{{Name: "volume2"}}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-0")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 3)
	assert.Len(t, conflictMap, 1)
	assert.Contains(t, conflictMap, "volume2")

	newName := conflictMap["volume2"]
	assert.True(t, strings.HasPrefix(newName, common.Fluid))
	assert.Equal(t, newName, resultVolumes[2].Name)
}

func TestAppendVolumes_MultipleConflicts(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "vol1"}, {Name: "vol2-0"}, {Name: "vol3-0"}}
	volumesToAdd := []corev1.Volume{{Name: "vol2"}, {Name: "vol3"}}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-0")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 5)
	assert.Len(t, conflictMap, 2)
	assert.Contains(t, conflictMap, "vol2")
	assert.Contains(t, conflictMap, "vol3")
}

func TestAppendVolumes_ConflictMappingsCorrect(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "data-volume-0"}}
	volumesToAdd := []corev1.Volume{{Name: "data-volume"}}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-0")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 2)
	assert.NotEqual(t, "data-volume-0", conflictMap["data-volume"])
	assert.NotEqual(t, "data-volume", conflictMap["data-volume"])
}

func TestAppendVolumes_SuffixDash1(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "volume1"}}
	volumesToAdd := []corev1.Volume{{Name: "volume2"}}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-1")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 2)
	assert.Len(t, conflictMap, 1)
	assert.Equal(t, "volume2-1", conflictMap["volume2"])
	assert.Equal(t, "volume2-1", resultVolumes[1].Name)
}

func TestAppendVolumes_CustomSuffix(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "volume1"}}
	volumesToAdd := []corev1.Volume{{Name: "volume2"}}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-custom")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 2)
	assert.Len(t, conflictMap, 1)
	assert.Equal(t, "volume2-custom", conflictMap["volume2"])
	assert.Equal(t, "volume2-custom", resultVolumes[1].Name)
}

func TestAppendVolumes_PreservesVolumeProperties(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{{Name: "vol1"}}
	volumesToAdd := []corev1.Volume{
		{
			Name:         "vol2",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		},
	}

	_, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-0")

	assert.NoError(t, err)
	assert.NotNil(t, resultVolumes[1].VolumeSource.EmptyDir)
}

// ---- randomizeNewVolumeName tests ----

func TestRandomizeNewVolumeName_NoConflict(t *testing.T) {
	injector := newTestInjector(t)
	existingNames := []string{"volume1", "volume2"}

	newName, err := injector.randomizeNewVolumeName("new-volume", existingNames)

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(newName, common.Fluid))
	assert.NotEqual(t, "new-volume", newName)
	assert.NotContains(t, existingNames, newName)
}

func TestRandomizeNewVolumeName_FirstAttemptConflict(t *testing.T) {
	injector := newTestInjector(t)
	existingNames := []string{"volume1", "volume2"}

	newName, err := injector.randomizeNewVolumeName("test-volume", existingNames)

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(newName, common.Fluid))
	assert.NotContains(t, existingNames, newName)
}

func TestRandomizeNewVolumeName_MultipleRetries(t *testing.T) {
	injector := newTestInjector(t)
	existingNames := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		existingNames = append(existingNames, "vol-"+utils.RandomAlphaNumberString(3))
	}

	newName, err := injector.randomizeNewVolumeName("conflict-volume", existingNames)

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(newName, common.Fluid))
	assert.NotContains(t, existingNames, newName)
}

func TestRandomizeNewVolumeName_PrefixReplaced(t *testing.T) {
	injector := newTestInjector(t)

	newName, err := injector.randomizeNewVolumeName("my-prefix-volume", []string{})

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(newName, common.Fluid))
	assert.False(t, strings.HasPrefix(newName, "my-prefix"))
}

func TestRandomizeNewVolumeName_DifferentNamesOnSubsequentCalls(t *testing.T) {
	injector := newTestInjector(t)
	existingNames := []string{}

	name1, err1 := injector.randomizeNewVolumeName("test-volume", existingNames)
	existingNames = append(existingNames, name1)
	name2, err2 := injector.randomizeNewVolumeName("test-volume", existingNames)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, name1, name2)
}

func TestRandomizeNewVolumeName_EmptyInput(t *testing.T) {
	injector := newTestInjector(t)

	newName, err := injector.randomizeNewVolumeName("", []string{"vol1", "vol2"})

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(newName, common.Fluid))
}

func TestRandomizeNewVolumeName_EmptyExistingNames(t *testing.T) {
	injector := newTestInjector(t)

	newName, err := injector.randomizeNewVolumeName("test-volume", []string{})

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(newName, common.Fluid))
}

func TestRandomizeNewVolumeName_ManyExistingNames(t *testing.T) {
	injector := newTestInjector(t)
	existingNames := make([]string, 0, 50)
	for i := 0; i < 50; i++ {
		existingNames = append(existingNames, "volume-"+string(rune(i)))
	}

	newName, err := injector.randomizeNewVolumeName("new-volume", existingNames)

	assert.NoError(t, err)
	assert.NotEmpty(t, newName)
	assert.NotContains(t, existingNames, newName)
}

// ---- Integration: appendVolumes with randomizeNewVolumeName ----

func TestAppendVolumes_Integration_ComplexConflictResolution(t *testing.T) {
	injector := newTestInjector(t)
	existingVolumes := []corev1.Volume{
		{Name: "vol-a"},
		{Name: "vol-b-0"},
		{Name: "vol-c-1"},
	}
	volumesToAdd := []corev1.Volume{
		{Name: "vol-b"},
		{Name: "vol-c"},
		{Name: "vol-d"},
	}

	conflictMap, resultVolumes, err := injector.appendVolumes(existingVolumes, volumesToAdd, "-0")

	assert.NoError(t, err)
	assert.Len(t, resultVolumes, 6)
	assert.Len(t, conflictMap, 3)

	assert.Contains(t, conflictMap, "vol-b")
	assert.True(t, strings.HasPrefix(conflictMap["vol-b"], common.Fluid))

	assert.Contains(t, conflictMap, "vol-c")
	assert.Equal(t, "vol-c-0", conflictMap["vol-c"])

	assert.Contains(t, conflictMap, "vol-d")
	assert.Equal(t, "vol-d-0", conflictMap["vol-d"])

	// Verify all names are unique
	nameSet := make(map[string]bool)
	for _, vol := range resultVolumes {
		assert.False(t, nameSet[vol.Name], "Duplicate volume name: "+vol.Name)
		nameSet[vol.Name] = true
	}
}
