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

package alluxio

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	runtimeOpts "github.com/fluid-cloudnative/fluid/pkg/utils/runtimes/options"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// SyncRuntime syncs the runtime spec
func (e *AlluxioEngine) SyncRuntime(ctx runtime.ReconcileRequestContext) (changed bool, err error) {
	if runtimeOpts.ShouldSkipSyncingRuntime() {
		e.Log.V(1).Info("Skipping runtime sync due to CONTROLLER_SKIP_SYNCING_RUNTIME being enabled")
		return
	}

	alluxioRuntime, err := e.getRuntime()
	if err != nil {
		return false, err
	}

	_, _, _, commonImagePullSecrets := e.parseRuntimeImage(
		alluxioRuntime.Spec.AlluxioVersion.Image,
		alluxioRuntime.Spec.AlluxioVersion.ImageTag,
		alluxioRuntime.Spec.AlluxioVersion.ImagePullPolicy,
		alluxioRuntime.Spec.ImagePullSecrets,
	)
	masterImagePullSecrets := alluxioRuntime.Spec.Master.ImagePullSecrets
	if len(masterImagePullSecrets) == 0 {
		masterImagePullSecrets = commonImagePullSecrets
	}
	workerImagePullSecrets := alluxioRuntime.Spec.Worker.ImagePullSecrets
	if len(workerImagePullSecrets) == 0 {
		workerImagePullSecrets = commonImagePullSecrets
	}
	_, _, _, fuseImagePullSecrets := e.parseFuseImage(
		alluxioRuntime.Spec.Fuse.Image,
		alluxioRuntime.Spec.Fuse.ImageTag,
		alluxioRuntime.Spec.Fuse.ImagePullPolicy,
		alluxioRuntime.Spec.Fuse.ImagePullSecrets,
	)
	if len(fuseImagePullSecrets) == 0 {
		fuseImagePullSecrets = commonImagePullSecrets
	}

	masterChanged, err := e.syncStatefulSetImagePullSecrets(e.getMasterName(), masterImagePullSecrets)
	changed = changed || masterChanged
	if err != nil {
		return changed, fmt.Errorf("failed to sync master imagePullSecrets: %w", err)
	}

	workerChanged, err := e.syncStatefulSetImagePullSecrets(e.getWorkerName(), workerImagePullSecrets)
	changed = changed || workerChanged
	if err != nil {
		return changed, fmt.Errorf("failed to sync worker imagePullSecrets: %w", err)
	}

	fuseChanged, err := e.syncDaemonSetImagePullSecrets(e.getFuseName(), fuseImagePullSecrets)
	changed = changed || fuseChanged
	if err != nil {
		return changed, fmt.Errorf("failed to sync fuse imagePullSecrets: %w", err)
	}

	return changed, nil
}

func (e *AlluxioEngine) syncStatefulSetImagePullSecrets(name string, imagePullSecrets []corev1.LocalObjectReference) (changed bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		statefulSet := &appsv1.StatefulSet{}
		if err := e.Get(context.TODO(), types.NamespacedName{Namespace: e.namespace, Name: name}, statefulSet); err != nil {
			return err
		}

		if reflect.DeepEqual(statefulSet.Spec.Template.Spec.ImagePullSecrets, imagePullSecrets) {
			return nil
		}

		statefulSetToUpdate := statefulSet.DeepCopy()
		statefulSetToUpdate.Spec.Template.Spec.ImagePullSecrets = append([]corev1.LocalObjectReference(nil), imagePullSecrets...)
		if err := e.Update(context.TODO(), statefulSetToUpdate); err != nil {
			return err
		}

		changed = true
		return nil
	})
	return changed, err
}

func (e *AlluxioEngine) syncDaemonSetImagePullSecrets(name string, imagePullSecrets []corev1.LocalObjectReference) (changed bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		daemonSet := &appsv1.DaemonSet{}
		if err := e.Get(context.TODO(), types.NamespacedName{Namespace: e.namespace, Name: name}, daemonSet); err != nil {
			return err
		}

		if reflect.DeepEqual(daemonSet.Spec.Template.Spec.ImagePullSecrets, imagePullSecrets) {
			return nil
		}

		daemonSetToUpdate := daemonSet.DeepCopy()
		daemonSetToUpdate.Spec.Template.Spec.ImagePullSecrets = append([]corev1.LocalObjectReference(nil), imagePullSecrets...)
		if err := e.Update(context.TODO(), daemonSetToUpdate); err != nil {
			return err
		}

		changed = true
		return nil
	})
	return changed, err
}
