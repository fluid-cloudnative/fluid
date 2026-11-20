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

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Binding a CacheRuntime to a snapshot of its CacheRuntimeClass
//
// The first time a CacheRuntime is set up, the CacheRuntimeClass it names is copied into an
// immutable configmap the runtime controls. From then on the runtime reads its class only from
// that copy, so editing or deleting a CacheRuntimeClass affects only the CacheRuntimes created
// afterwards.
//
// The whole class is copied, not just the component templates. The images, their options, the
// commands Fluid runs inside the components and the data operation specs are written for one
// another: following part of the class live would hand new commands to components still running
// old images. Following all of it live is no better. The class is cluster-scoped and the bundled
// AdvancedStatefulSet controller restarts the containers for an image change and recreates the
// Pods for a resources change, so a single class edit would rebuild the cache of every
// CacheRuntime using that class.

// runtimeClassSnapshotKey is the data key of the snapshot configmap that holds the class as JSON.
const runtimeClassSnapshotKey = "runtimeclass.json"

const (
	snapshotWaitInterval = 100 * time.Millisecond
	snapshotWaitTimeout  = 3 * time.Second
)

// getBoundRuntimeClass returns the CacheRuntimeClass the runtime is bound to, binding it to a
// snapshot of its current class first when it has none.
//
// The client reads through the informer cache, which does not see a snapshot created moments
// ago, so the caller passes the returned class down instead of reading the snapshot again in the
// same reconcile.
func (e *CacheEngine) getBoundRuntimeClass(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*datav1alpha1.CacheRuntimeClass, error) {
	snapshot, err := e.getRuntimeClassSnapshot(ctx, runtime)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		snapshot, err = e.bindRuntimeClass(ctx, runtime)
		if err != nil {
			return nil, err
		}
	}
	return e.loadRuntimeClassSnapshot(runtime, snapshot)
}

// getBoundRuntimeClassReadOnly returns the CacheRuntimeClass the runtime is bound to without
// changing anything. Data operations run in the dataset controller and use it: only the runtime
// controller binds a runtime or replaces its snapshot.
//
// It never falls back to the current class. A runtime without a snapshot of its own is one whose
// snapshot was deleted to rebind it, or one this controller's cache shows out of date; reading the
// current class then could run a data operation from a class its components do not run yet. An
// error is returned instead, and the data operation is retried once the runtime is bound.
func (e *CacheEngine) getBoundRuntimeClassReadOnly(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*datav1alpha1.CacheRuntimeClass, error) {
	snapshot, err := e.readRuntimeClassSnapshot(ctx, runtime)
	if err != nil {
		return nil, err
	}
	name := common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name)
	if snapshot == nil {
		return nil, fmt.Errorf("CacheRuntime %s/%s is not bound to a CacheRuntimeClass snapshot yet, retrying",
			runtime.Namespace, runtime.Name)
	}
	if !metav1.IsControlledBy(snapshot, runtime) {
		return nil, fmt.Errorf("CacheRuntimeClass snapshot %s/%s is not controlled by CacheRuntime %s (uid %s) yet, retrying",
			runtime.Namespace, name, runtime.Name, runtime.UID)
	}
	return e.loadRuntimeClassSnapshot(runtime, snapshot)
}

// readRuntimeClassSnapshot returns the configmap holding the runtime's snapshot, whoever controls
// it, or nil when there is none.
func (e *CacheEngine) readRuntimeClassSnapshot(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*corev1.ConfigMap, error) {
	return kubeclient.GetConfigmapByNameWithContext(ctx, e.Client,
		common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name), runtime.Namespace)
}

// getRuntimeClassSnapshot returns the runtime's snapshot configmap, or nil when it has none. Only
// the runtime controller calls it.
//
// A configmap of that name controlled by anything else, such as an earlier CacheRuntime of the
// same name the garbage collector has not removed yet, does not describe this runtime. It is
// deleted and nil returned, so the runtime is bound to a snapshot of its own.
//
// The ownership check runs on the informer cache, which can lag behind: the configmap it shows
// may already have been replaced by this runtime's own snapshot. The delete is therefore guarded
// by the UID that was checked, so it can only remove the configmap that was seen.
func (e *CacheEngine) getRuntimeClassSnapshot(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*corev1.ConfigMap, error) {
	snapshot, err := e.readRuntimeClassSnapshot(ctx, runtime)
	if err != nil || snapshot == nil {
		return nil, err
	}
	if metav1.IsControlledBy(snapshot, runtime) {
		return snapshot, nil
	}

	name := snapshot.Name
	e.Log.Info("Deleting a CacheRuntimeClass snapshot the CacheRuntime does not control", "snapshot", name)
	err = e.Client.Delete(ctx, snapshot, client.Preconditions{UID: ptr.To(snapshot.UID)})
	switch {
	case err == nil, apierrors.IsNotFound(err):
		return nil, nil
	case apierrors.IsConflict(err):
		// Replaced since the cache was filled; the next reconcile reads the current one.
		return nil, fmt.Errorf("CacheRuntimeClass snapshot %s/%s changed while it was being replaced, retrying: %w",
			runtime.Namespace, name, err)
	default:
		return nil, err
	}
}

// loadRuntimeClassSnapshot decodes the CacheRuntimeClass held by the runtime's snapshot.
func (e *CacheEngine) loadRuntimeClassSnapshot(runtime *datav1alpha1.CacheRuntime, snapshot *corev1.ConfigMap) (*datav1alpha1.CacheRuntimeClass, error) {
	runtimeClass, err := decodeRuntimeClassSnapshot(snapshot)
	if err != nil {
		// The snapshot is never rewritten: it holds the class the components were created
		// from, and a guess would move them. Deleting it binds the runtime to its current
		// class instead, which the user decides to do.
		err = fmt.Errorf("CacheRuntimeClass snapshot %s/%s is invalid: %w", snapshot.Namespace, snapshot.Name, err)
		e.recordWarning(runtime, common.RuntimeClassSnapshotInvalid,
			"%v; delete the configmap to bind the CacheRuntime to CacheRuntimeClass %q again", err, runtime.Spec.RuntimeClassName)
		return nil, err
	}

	if bound := snapshot.Annotations[common.AnnotationCacheRuntimeClassName]; bound != "" && bound != runtime.Spec.RuntimeClassName {
		e.recordWarning(runtime, common.RuntimeClassChanged,
			"runtimeClassName changed to %q, but the CacheRuntime stays bound to the snapshot of CacheRuntimeClass %q taken when it was set up; recreate the CacheRuntime to use %q",
			runtime.Spec.RuntimeClassName, bound, runtime.Spec.RuntimeClassName)
	}
	return runtimeClass, nil
}

func decodeRuntimeClassSnapshot(snapshot *corev1.ConfigMap) (*datav1alpha1.CacheRuntimeClass, error) {
	raw, ok := snapshot.Data[runtimeClassSnapshotKey]
	if !ok {
		return nil, fmt.Errorf("key %q not found", runtimeClassSnapshotKey)
	}
	runtimeClass := &datav1alpha1.CacheRuntimeClass{}
	if err := json.Unmarshal([]byte(raw), runtimeClass); err != nil {
		return nil, err
	}
	if err := validateRuntimeClassTopology(runtimeClass); err != nil {
		return nil, err
	}
	return runtimeClass, nil
}

// bindRuntimeClass binds the runtime to a snapshot of its current CacheRuntimeClass and returns
// the snapshot configmap.
func (e *CacheEngine) bindRuntimeClass(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*corev1.ConfigMap, error) {
	runtimeClass, err := e.getRuntimeClass(runtime.Spec.RuntimeClassName)
	if err != nil {
		return nil, err
	}

	snapshotClass := runtimeClass.DeepCopy()
	snapshotClass.TypeMeta = metav1.TypeMeta{APIVersion: datav1alpha1.GroupVersion.String(), Kind: "CacheRuntimeClass"}
	snapshotClass.ObjectMeta = metav1.ObjectMeta{Name: runtimeClass.Name}

	data, err := json.Marshal(snapshotClass)
	if err != nil {
		return nil, err
	}
	snapshot := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name),
			Namespace: runtime.Namespace,
			Labels:    map[string]string{common.LabelCacheRuntimeName: runtime.Name},
			Annotations: map[string]string{
				common.AnnotationCacheRuntimeClassName: runtimeClass.Name,
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         datav1alpha1.GroupVersion.String(),
				Kind:               datav1alpha1.CacheRuntimeKind,
				Name:               runtime.Name,
				UID:                runtime.UID,
				Controller:         ptr.To(true),
				BlockOwnerDeletion: ptr.To(true),
			}},
		},
		Data:      map[string]string{runtimeClassSnapshotKey: string(data)},
		Immutable: ptr.To(true),
	}
	if err := e.Client.Create(ctx, snapshot); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// An earlier reconcile bound the runtime and the informer cache has not caught
			// up yet. The snapshot that exists wins.
			return e.waitForRuntimeClassSnapshot(ctx, runtime)
		}
		return nil, err
	}

	e.Log.Info("Bound the CacheRuntime to a snapshot of its CacheRuntimeClass",
		"runtimeClass", runtimeClass.Name, "snapshot", snapshot.Name)
	return snapshot, nil
}

// waitForRuntimeClassSnapshot waits briefly for an existing snapshot to reach the informer
// cache the client reads from.
func (e *CacheEngine) waitForRuntimeClassSnapshot(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*corev1.ConfigMap, error) {
	var snapshot *corev1.ConfigMap
	err := wait.PollUntilContextTimeout(ctx, snapshotWaitInterval, snapshotWaitTimeout, true, func(ctx context.Context) (bool, error) {
		var err error
		snapshot, err = e.getRuntimeClassSnapshot(ctx, runtime)
		return snapshot != nil, err
	})
	if err != nil {
		return nil, fmt.Errorf("CacheRuntimeClass snapshot %s/%s exists but cannot be read yet: %w",
			runtime.Namespace, common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name), err)
	}
	return snapshot, nil
}

// recordWarning emits a Warning event on the runtime when the engine has a recorder.
func (e *CacheEngine) recordWarning(runtime *datav1alpha1.CacheRuntime, reason, messageFmt string, args ...interface{}) {
	if runtime == nil || e.Recorder == nil {
		return
	}
	e.Recorder.Eventf(runtime, corev1.EventTypeWarning, reason, messageFmt, args...)
}
