/*
Copyright 2026 The Fluid Authors.

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

package juicefs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func pullSecrets(names ...string) []corev1.LocalObjectReference {
	var out []corev1.LocalObjectReference
	for _, n := range names {
		out = append(out, corev1.LocalObjectReference{Name: n})
	}
	return out
}

func TestSyncImagePullSecrets(t *testing.T) {
	cases := []struct {
		name string
		// old and latest values: global, worker, fuse
		oldGlobal, oldWorker, oldFuse []corev1.LocalObjectReference
		newGlobal, newWorker, newFuse []corev1.LocalObjectReference
		wantWorker, wantFuse          []corev1.LocalObjectReference
		// live templates, when they differ from what the old value implies
		liveWorker, liveFuse                  []corev1.LocalObjectReference
		liveSet                               bool
		workerChanged, fuseChanged, fuseBumps bool
	}{
		{
			name:       "runtime-wide secrets added reach both components",
			newGlobal:  pullSecrets("a"),
			wantWorker: pullSecrets("a"), wantFuse: pullSecrets("a"),
			workerChanged: true, fuseChanged: true, fuseBumps: true,
		},
		{
			name:      "worker override removed falls back to the runtime-wide secrets",
			oldGlobal: pullSecrets("g"), oldWorker: pullSecrets("w"),
			newGlobal:  pullSecrets("g"),
			wantWorker: pullSecrets("g"), wantFuse: pullSecrets("g"),
			workerChanged: true,
		},
		{
			name:      "fuse override replaced only changes fuse",
			oldGlobal: pullSecrets("g"), oldFuse: pullSecrets("f1"),
			newGlobal: pullSecrets("g"), newFuse: pullSecrets("f2"),
			wantWorker: pullSecrets("g"), wantFuse: pullSecrets("f2"),
			fuseChanged: true, fuseBumps: true,
		},
		{
			name:       "secrets removed entirely",
			oldGlobal:  pullSecrets("g"),
			wantWorker: nil, wantFuse: nil,
			workerChanged: true, fuseChanged: true, fuseBumps: true,
		},
		{
			name:      "template left on the old secrets by a skipped sync catches up",
			oldGlobal: pullSecrets("b"), newGlobal: pullSecrets("b"),
			liveWorker: pullSecrets("a"), liveFuse: pullSecrets("a"), liveSet: true,
			wantWorker: pullSecrets("b"), wantFuse: pullSecrets("b"),
			workerChanged: true, fuseChanged: true, fuseBumps: true,
		},
		{
			name:      "override added while the runtime-wide secrets rotate keeps the template as is",
			oldGlobal: pullSecrets("a"), newGlobal: pullSecrets("b"), newWorker: pullSecrets("a"),
			wantWorker: pullSecrets("a"), wantFuse: pullSecrets("b"),
			fuseChanged: true, fuseBumps: true,
		},
		{
			name:      "no change",
			oldGlobal: pullSecrets("g"), newGlobal: pullSecrets("g"),
			wantWorker: pullSecrets("g"), wantFuse: pullSecrets("g"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			engine := &JuiceFSEngine{name: "test", namespace: "default"}
			runtime := &datav1alpha1.JuiceFSRuntime{}

			oldValue := constructBaseRuntimeValue()
			latestValue := constructBaseRuntimeValue()
			oldValue.ImagePullSecrets, oldValue.Worker.ImagePullSecrets, oldValue.Fuse.ImagePullSecrets = c.oldGlobal, c.oldWorker, c.oldFuse
			latestValue.ImagePullSecrets, latestValue.Worker.ImagePullSecrets, latestValue.Fuse.ImagePullSecrets = c.newGlobal, c.newWorker, c.newFuse

			worker := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: JuiceFSWorkerContainerName}}}}}}
			worker.Spec.Template.Spec.ImagePullSecrets = effectiveImagePullSecrets(c.oldWorker, c.oldGlobal)
			if c.liveSet {
				worker.Spec.Template.Spec.ImagePullSecrets = c.liveWorker
			}
			gotWorkerChanged := engine.checkAndSetWorkerChanges(oldValue, latestValue, runtime, worker)
			assert.Equal(t, c.workerChanged, gotWorkerChanged)
			assert.Equal(t, c.wantWorker, worker.Spec.Template.Spec.ImagePullSecrets)

			fuse := constructBaseFuseDaemonset()
			fuse.Spec.Template.Spec.ImagePullSecrets = effectiveImagePullSecrets(c.oldFuse, c.oldGlobal)
			if c.liveSet {
				fuse.Spec.Template.Spec.ImagePullSecrets = c.liveFuse
			}
			gotFuseChanged, gotBump := engine.checkAndSetFuseChanges(oldValue, latestValue, runtime, fuse)
			// the shared fuse fixture can report changes for other fields, so only the positive case is checked
			if c.fuseChanged {
				assert.True(t, gotFuseChanged)
			}
			assert.Equal(t, c.fuseBumps, gotBump)
			assert.Equal(t, c.wantFuse, fuse.Spec.Template.Spec.ImagePullSecrets)

			// the overrides are recorded whether or not the template changed
			assert.Equal(t, c.newWorker, oldValue.Worker.ImagePullSecrets)
			assert.Equal(t, c.newFuse, oldValue.Fuse.ImagePullSecrets)
		})
	}
}

// Rotating the runtime-wide secrets from a to b while adding a worker override a, then
// removing the override, has to leave the worker on b.
func TestSyncImagePullSecretsRotationThenOverrideRemoval(t *testing.T) {
	engine := &JuiceFSEngine{name: "test", namespace: "default"}
	runtime := &datav1alpha1.JuiceFSRuntime{}
	worker := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
		Containers:       []corev1.Container{{Name: JuiceFSWorkerContainerName}},
		ImagePullSecrets: pullSecrets("a"),
	}}}}

	saved := constructBaseRuntimeValue()
	saved.ImagePullSecrets = pullSecrets("a")

	// rotation plus override: the effective list stays a
	latest := constructBaseRuntimeValue()
	latest.ImagePullSecrets, latest.Worker.ImagePullSecrets = pullSecrets("b"), pullSecrets("a")
	assert.False(t, engine.checkAndSetWorkerChanges(saved, latest, runtime, worker))
	saved.ImagePullSecrets = latest.ImagePullSecrets
	assert.Equal(t, pullSecrets("a"), saved.Worker.ImagePullSecrets)

	// override removed: the worker falls back to b
	latest = constructBaseRuntimeValue()
	latest.ImagePullSecrets = pullSecrets("b")
	assert.True(t, engine.checkAndSetWorkerChanges(saved, latest, runtime, worker))
	assert.Equal(t, pullSecrets("b"), worker.Spec.Template.Spec.ImagePullSecrets)
	assert.Nil(t, saved.Worker.ImagePullSecrets)
}
