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
	"slices"
	"strings"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
)

// runtimeClassSnapshotKey is the key of the snapshot configmap holding the CacheRuntimeClass.
const runtimeClassSnapshotKey = "runtimeclass.json"

// A CacheRuntime is bound to a snapshot of its CacheRuntimeClass, taken when the runtime is
// first set up and kept in an immutable configmap the runtime owns. Every later read of the
// class goes through the snapshot, so editing or deleting a CacheRuntimeClass only affects
// the CacheRuntimes created afterwards.
//
// The class is a versioned bundle: the component images, their options, the commands Fluid
// executes in the components and the data operation specs are written for one another. Were
// part of it followed live, a class upgrade would hand new commands and options to components
// still running the old images. Following all of it live is no better: the bundled
// AdvancedStatefulSet controller recreates Pods for any resources change and restarts
// containers for an image change, so one edit of a cluster-scoped class would rebuild the
// caches of every CacheRuntime using it.

// getBoundRuntimeClass returns the CacheRuntimeClass the runtime is bound to, binding the
// runtime to a snapshot of its current class first when it has none yet. It also returns the
// components whose adoption is pending, see holdAdoptionPending.
//
// Both come from the snapshot read, or created, here. The client reads through the informer
// cache, which does not see a snapshot created moments ago, so the caller must not read the
// snapshot again within the same reconcile.
func (e *CacheEngine) getBoundRuntimeClass(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (
	runtimeClass *datav1alpha1.CacheRuntimeClass, adoptionPending map[common.ComponentType]bool, err error) {
	snapshot, err := e.getRuntimeClassSnapshot(ctx, runtime)
	if err != nil {
		return nil, nil, err
	}
	if snapshot == nil {
		if snapshot, err = e.bindRuntimeClass(ctx, runtime); err != nil {
			return nil, nil, err
		}
	}
	if runtimeClass, err = e.loadRuntimeClassSnapshot(runtime, snapshot); err != nil {
		return nil, nil, err
	}
	return runtimeClass, adoptionPendingComponents(snapshot), nil
}

// getBoundRuntimeClassReadOnly returns the CacheRuntimeClass the runtime is bound to without
// binding it. It is used outside the runtime controller, by data operations: a runtime set up
// before snapshots existed has none until the runtime controller adopts it, which it does on
// the runtime's next reconcile, and until then the class is read as it was before.
func (e *CacheEngine) getBoundRuntimeClassReadOnly(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*datav1alpha1.CacheRuntimeClass, error) {
	snapshot, err := e.getRuntimeClassSnapshot(ctx, runtime)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return e.loadRuntimeClassSnapshot(runtime, snapshot)
	}
	e.Log.Info("CacheRuntime is not bound to a CacheRuntimeClass snapshot yet, reading the current class",
		"runtimeClass", runtime.Spec.RuntimeClassName)
	return e.getRuntimeClass(runtime.Spec.RuntimeClassName)
}

// getRuntimeClassSnapshot returns the snapshot configmap of the runtime, or nil when there is
// none. A snapshot controlled by another object, such as an earlier CacheRuntime of the same
// name the garbage collector has not removed yet, describes something else: it is deleted and
// nil returned, so that the runtime gets a snapshot of its own.
func (e *CacheEngine) getRuntimeClassSnapshot(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*corev1.ConfigMap, error) {
	name := common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name)
	snapshot, err := kubeclient.GetConfigmapByNameWithContext(ctx, e.Client, name, runtime.Namespace)
	if err != nil || snapshot == nil {
		return nil, err
	}
	if metav1.IsControlledBy(snapshot, runtime) {
		return snapshot, nil
	}

	e.Log.Info("Deleting a CacheRuntimeClass snapshot that does not belong to the CacheRuntime", "snapshot", name)
	if err := e.Client.Delete(ctx, snapshot); err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	return nil, nil
}

// loadRuntimeClassSnapshot decodes the CacheRuntimeClass held by a snapshot configmap.
func (e *CacheEngine) loadRuntimeClassSnapshot(runtime *datav1alpha1.CacheRuntime, snapshot *corev1.ConfigMap) (*datav1alpha1.CacheRuntimeClass, error) {
	runtimeClass, err := decodeRuntimeClassSnapshot(snapshot)
	if err != nil {
		// Never overwritten in place: the snapshot holds the defaults the components were
		// created from, and a guess would move them. Deleting the configmap binds the runtime
		// to its current class again, adopting what the components run today.
		err = fmt.Errorf("the CacheRuntimeClass snapshot %s/%s is invalid: %w", snapshot.Namespace, snapshot.Name, err)
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
	raw, found := snapshot.Data[runtimeClassSnapshotKey]
	if !found {
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

// bindRuntimeClass takes a snapshot of the runtime's current CacheRuntimeClass, binds the
// runtime to it and returns the snapshot configmap. A runtime set up before snapshots existed
// has components already: their defaults are adopted from what they run (see
// adoptExistingComponents), so binding never changes a running component.
func (e *CacheEngine) bindRuntimeClass(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*corev1.ConfigMap, error) {
	runtimeClass, err := e.getRuntimeClass(runtime.Spec.RuntimeClassName)
	if err != nil {
		return nil, err
	}

	snapshotClass := runtimeClass.DeepCopy()
	snapshotClass.TypeMeta = metav1.TypeMeta{APIVersion: datav1alpha1.GroupVersion.String(), Kind: "CacheRuntimeClass"}
	snapshotClass.ObjectMeta = metav1.ObjectMeta{Name: runtimeClass.Name}

	pending, err := e.adoptExistingComponents(ctx, runtime, snapshotClass)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(snapshotClass)
	if err != nil {
		return nil, err
	}
	snapshot := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name),
			Namespace: runtime.Namespace,
			Labels: map[string]string{
				common.LabelCacheRuntimeName: runtime.Name,
			},
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
	if len(pending) > 0 {
		snapshot.Annotations[common.AnnotationCacheRuntimeAdoptionPending] = joinComponentTypes(pending)
	}

	if err := e.Client.Create(ctx, snapshot); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Bound by an earlier reconcile the informer cache has not caught up with. The
			// snapshot that exists wins.
			return e.waitForRuntimeClassSnapshot(ctx, runtime)
		}
		return nil, err
	}

	e.Log.Info("Bound the CacheRuntime to a snapshot of its CacheRuntimeClass",
		"runtimeClass", runtimeClass.Name, "snapshot", snapshot.Name, "adoptionPending", pending)
	for _, componentType := range pending {
		e.recordWarning(runtime, common.RuntimeAdoptionPending, "%s", adoptionPendingMessage(componentType, snapshot.Name))
	}

	return snapshot, nil
}

// waitForRuntimeClassSnapshot waits briefly for a snapshot that exists to show up in the
// informer cache the client reads from.
func (e *CacheEngine) waitForRuntimeClassSnapshot(ctx context.Context, runtime *datav1alpha1.CacheRuntime) (*corev1.ConfigMap, error) {
	var snapshot *corev1.ConfigMap
	err := wait.PollUntilContextTimeout(ctx, 100*time.Millisecond, 3*time.Second, true, func(ctx context.Context) (bool, error) {
		var err error
		snapshot, err = e.getRuntimeClassSnapshot(ctx, runtime)
		return snapshot != nil, err
	})
	if err != nil {
		return nil, fmt.Errorf("CacheRuntimeClass snapshot %s/%s exists but cannot be read yet: %w", runtime.Namespace,
			common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name), err)
	}
	return snapshot, nil
}

// adoptExistingComponents replaces, in snapshotClass, the template resources and image of the
// master and worker that already exist with the defaults adopted from their workloads (see
// adoptComponentBaseline). It returns the components whose workload would still change were
// the runtime's spec synced on top of the adopted defaults. A component that does not exist
// yet keeps the class template: it is created from the snapshot.
//
// The client is not adopted: it runs as a DaemonSet that is never synced after creation.
func (e *CacheEngine) adoptExistingComponents(ctx context.Context, runtime *datav1alpha1.CacheRuntime,
	snapshotClass *datav1alpha1.CacheRuntimeClass) (pending []common.ComponentType, err error) {
	components := []struct {
		componentType common.ComponentType
		definition    *datav1alpha1.RuntimeComponentDefinition
		spec          datav1alpha1.RuntimeComponentCommonSpec
	}{
		{common.ComponentTypeMaster, snapshotClass.Topology.Master, runtime.Spec.Master.RuntimeComponentCommonSpec},
		{common.ComponentTypeWorker, snapshotClass.Topology.Worker, runtime.Spec.Worker.RuntimeComponentCommonSpec},
	}

	for _, c := range components {
		if c.definition == nil || len(c.definition.Template.Spec.Containers) == 0 {
			continue
		}
		identity := &common.ComponentIdentity{
			Name:      common.GetCacheComponentName(e.name, c.componentType),
			Namespace: e.namespace,
		}
		podSpec, err := component.NewComponentHelper(c.componentType, e.Client).GetPodSpec(ctx, identity)
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if len(podSpec.Containers) == 0 {
			continue
		}

		baseline := adoptComponentBaseline(c.componentType, c.spec, podSpec)
		c.definition.Template.Spec.Containers[0].Resources = baseline.resources
		c.definition.Template.Spec.Containers[0].Image = baseline.image
		if !baseline.applied {
			pending = append(pending, c.componentType)
		}
	}

	return pending, nil
}

// adoptedBaseline is the class template default of a component, recovered from its workload.
type adoptedBaseline struct {
	resources corev1.ResourceRequirements
	image     string
	// applied is true when syncing the runtime's spec on top of the baseline leaves the
	// workload exactly as it is.
	applied bool
}

// adoptComponentBaseline recovers the template resources and image a component was created
// from, out of what its workload runs, without reading the CacheRuntimeClass: the class may
// have changed since, and reading it would let that change reach the component.
//
//   - resources: the workload's, minus the keys the runtime names, whose workload value came
//     from the runtime, and, for the worker, minus the tiered store memory quota charged on
//     top. Removing an override from such a runtime later clears the key rather than restoring
//     a template value nobody recorded.
//   - image: the workload's. It completes a partial runtimeVersion exactly as the template did
//     when the component was rendered or last synced.
func adoptComponentBaseline(componentType common.ComponentType, spec datav1alpha1.RuntimeComponentCommonSpec,
	podSpec *corev1.PodSpec) adoptedBaseline {
	container := podSpec.Containers[0]

	var quota resource.Quantity
	if componentType == common.ComponentTypeWorker {
		quota = chargedTieredStoreMemoryQuota(podSpec)
	}

	resources := *container.Resources.DeepCopy()
	removeNamedResources(&resources, spec.Resources)
	invertible := withoutTieredStoreMemoryQuota(&resources, quota)

	return adoptedBaseline{
		resources: resources,
		image:     container.Image,
		applied: invertible &&
			resourcesSyncedUnchanged(componentType, resources, spec.Resources, podSpec) &&
			imageSyncedUnchanged(spec.RuntimeVersion, container.Image, container.Image),
	}
}

// removeNamedResources deletes from resources every request, limit and claim that named names.
func removeNamedResources(resources *corev1.ResourceRequirements, named corev1.ResourceRequirements) {
	for name := range named.Requests {
		delete(resources.Requests, name)
	}
	for name := range named.Limits {
		delete(resources.Limits, name)
	}
	if len(named.Claims) == 0 {
		return
	}
	resources.Claims = slices.DeleteFunc(resources.Claims, func(claim corev1.ResourceClaim) bool {
		return slices.ContainsFunc(named.Claims, func(other corev1.ResourceClaim) bool { return other.Name == claim.Name })
	})
}

// withoutTieredStoreMemoryQuota takes the tiered store memory quota back out of the memory
// request and limit it was charged to, the inverse of withTieredStoreMemoryQuota. It reports
// false when a charged value is not above the quota, which withTieredStoreMemoryQuota cannot
// have produced; that value is then left as it is.
func withoutTieredStoreMemoryQuota(resources *corev1.ResourceRequirements, quota resource.Quantity) bool {
	if quota.IsZero() {
		return true
	}
	invertible := true
	for _, list := range []corev1.ResourceList{resources.Requests, resources.Limits} {
		current, exists := list[corev1.ResourceMemory]
		if !exists || current.IsZero() {
			continue
		}
		if current.Cmp(quota) <= 0 {
			invertible = false
			continue
		}
		current.Sub(quota)
		list[corev1.ResourceMemory] = current
	}
	return invertible
}

// resourcesSyncedUnchanged reports whether syncRuntimeSpec, resolving the runtime's resources
// over baseline, would leave the component's workload resources as they are. It follows the
// sync step by step: a conflicting merge is not synced at all, claims the workload does not
// declare are left out, and the worker is charged the tiered store memory quota.
func resourcesSyncedUnchanged(componentType common.ComponentType, baseline, runtimeResources corev1.ResourceRequirements,
	podSpec *corev1.PodSpec) bool {
	desired := mergeResourceRequirements(baseline, runtimeResources)
	if validateResourceRequirements(desired) != nil {
		return true
	}
	desired.Claims, _ = splitDeclaredClaims(desired.Claims, podSpec)
	if componentType == common.ComponentTypeWorker {
		desired = withTieredStoreMemoryQuota(desired, chargedTieredStoreMemoryQuota(podSpec))
	}
	return component.ResourcesEqual(podSpec.Containers[0].Resources, desired)
}

// imageSyncedUnchanged reports whether syncRuntimeSpec, completing runtimeVersion against
// templateImage, would leave the component running currentImage. An incomplete version is
// not synced at all.
func imageSyncedUnchanged(runtimeVersion datav1alpha1.VersionSpec, templateImage, currentImage string) bool {
	version := desiredComponentVersion(runtimeVersion, templateImage)
	if version.Image == "" || version.ImageTag == "" {
		return true
	}
	return version.Image+":"+version.ImageTag == currentImage
}

// adoptionPendingComponents returns the components whose adoption the snapshot lists as
// pending: syncing what the runtime declares would change them, which a component created
// before the runtime was bound must not undergo silently.
func adoptionPendingComponents(snapshot *corev1.ConfigMap) map[common.ComponentType]bool {
	pending := map[common.ComponentType]bool{}
	for _, componentType := range splitComponentTypes(snapshot.Annotations[common.AnnotationCacheRuntimeAdoptionPending]) {
		pending[componentType] = true
	}
	return pending
}

// holdAdoptionPending narrows the spec synced to a component whose adoption is pending to its
// replicas, and warns that its resources and image are held back. Once the workload already
// runs what spec asks for, nothing is held back any more and the component leaves the pending
// list.
func (e *CacheEngine) holdAdoptionPending(ctx context.Context, runtime *datav1alpha1.CacheRuntime, componentType common.ComponentType,
	podSpec *corev1.PodSpec, spec component.ComponentSpec) component.ComponentSpec {
	if componentSpecUnchanged(podSpec, spec) {
		if err := e.releaseAdoptionPending(ctx, runtime, componentType); err != nil {
			e.Log.Error(err, "Failed to remove the component from the adoption pending list", "component", componentType)
		}
		return spec
	}

	e.recordWarning(runtime, common.RuntimeAdoptionPending, "%s",
		adoptionPendingMessage(componentType, common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name)))
	return component.ComponentSpec{Replicas: spec.Replicas}
}

// componentSpecUnchanged reports whether syncing spec would leave the resources and the image
// of the workload's first container as they are.
func componentSpecUnchanged(podSpec *corev1.PodSpec, spec component.ComponentSpec) bool {
	if len(podSpec.Containers) == 0 {
		return true
	}
	container := podSpec.Containers[0]
	if spec.Resources != nil && !component.ResourcesEqual(container.Resources, *spec.Resources) {
		return false
	}
	if spec.Version.Image != "" && spec.Version.ImageTag != "" && spec.Version.Image+":"+spec.Version.ImageTag != container.Image {
		return false
	}
	return true
}

// releaseAdoptionPending removes componentType from the adoption pending list of the runtime's
// snapshot. Only the annotation changes: the data of the snapshot is immutable.
func (e *CacheEngine) releaseAdoptionPending(ctx context.Context, runtime *datav1alpha1.CacheRuntime, componentType common.ComponentType) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		snapshot, err := kubeclient.GetConfigmapByNameWithContext(ctx, e.Client,
			common.GetCacheRuntimeClassSnapshotConfigMapName(runtime.Name), runtime.Namespace)
		if err != nil || snapshot == nil {
			return err
		}
		pending := splitComponentTypes(snapshot.Annotations[common.AnnotationCacheRuntimeAdoptionPending])
		remaining := slices.DeleteFunc(slices.Clone(pending), func(other common.ComponentType) bool { return other == componentType })
		if len(remaining) == len(pending) {
			return nil
		}

		snapshotToUpdate := snapshot.DeepCopy()
		if len(remaining) == 0 {
			delete(snapshotToUpdate.Annotations, common.AnnotationCacheRuntimeAdoptionPending)
		} else {
			snapshotToUpdate.Annotations[common.AnnotationCacheRuntimeAdoptionPending] = joinComponentTypes(remaining)
		}
		if err := e.Client.Update(ctx, snapshotToUpdate); err != nil {
			return err
		}
		e.Log.Info("The component already runs what the CacheRuntime declares, adoption is complete", "component", componentType)
		return nil
	})
}

func adoptionPendingMessage(componentType common.ComponentType, snapshotName string) string {
	return fmt.Sprintf("%s was created before this CacheRuntime was bound to a snapshot of its CacheRuntimeClass, and the resources or image "+
		"the CacheRuntime declares differ from what it runs. Applying them would recreate its Pods (resources) or restart its "+
		"containers (image), so Fluid leaves it unchanged; remove %q from annotation %s of configmap %s to apply them",
		componentType, componentType, common.AnnotationCacheRuntimeAdoptionPending, snapshotName)
}

func joinComponentTypes(componentTypes []common.ComponentType) string {
	names := make([]string, 0, len(componentTypes))
	for _, componentType := range componentTypes {
		names = append(names, string(componentType))
	}
	return strings.Join(names, ",")
}

func splitComponentTypes(value string) []common.ComponentType {
	var componentTypes []common.ComponentType
	for _, name := range strings.Split(value, ",") {
		if name = strings.TrimSpace(name); name != "" {
			componentTypes = append(componentTypes, common.ComponentType(name))
		}
	}
	return componentTypes
}

// recordWarning emits a Warning event on the runtime, when there is a recorder.
func (e *CacheEngine) recordWarning(runtime *datav1alpha1.CacheRuntime, reason, messageFmt string, args ...interface{}) {
	if runtime == nil || e.Recorder == nil {
		return
	}
	e.Recorder.Eventf(runtime, corev1.EventTypeWarning, reason, messageFmt, args...)
}
