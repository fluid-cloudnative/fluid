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

// Package cacheruntime validates CacheRuntime admission requests.
package cacheruntime

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// UpdatableFieldsHint names the spec fields that are synced to the running
// components in place. It is appended to every rejection so the user learns
// the rule, not only the field that broke it.
const UpdatableFieldsHint = "only master/worker replicas, runtimeVersion and resources can be updated in place"

// ValidateSpecUpdate returns an error for every spec field that differs between
// oldSpec and newSpec and cannot take effect on a running CacheRuntime.
//
// The CacheRuntime controller syncs only the replicas, runtimeVersion and
// resources of the master and worker to their workloads (see
// AdvancedStatefulSetManager.SyncComponentSpec). The client is a DaemonSet and
// is never synced. Every other field is read once, when the component is
// created, so a later change would be accepted by the API server and then
// silently ignored, or, for tieredStore, applied halfway: the ConfigMap quota
// follows the spec while the tmpfs sizeLimit keeps its creation-time value.
//
// Only the spec is compared, so the controller's own writes to metadata and
// status are never affected.
func ValidateSpecUpdate(oldSpec, newSpec *datav1alpha1.CacheRuntimeSpec) field.ErrorList {
	root := field.NewPath("spec")
	var errs field.ErrorList

	errs = append(errs, requireUnchanged(root.Child("runtimeClassName"), oldSpec.RuntimeClassName, newSpec.RuntimeClassName)...)

	master := root.Child("master")
	errs = append(errs, validateComponentCommonSpecUpdate(master, &oldSpec.Master.RuntimeComponentCommonSpec, &newSpec.Master.RuntimeComponentCommonSpec, true)...)

	worker := root.Child("worker")
	errs = append(errs, validateComponentCommonSpecUpdate(worker, &oldSpec.Worker.RuntimeComponentCommonSpec, &newSpec.Worker.RuntimeComponentCommonSpec, true)...)
	errs = append(errs, requireUnchanged(worker.Child("tieredStore"), oldSpec.Worker.TieredStore, newSpec.Worker.TieredStore)...)

	client := root.Child("client")
	errs = append(errs, validateComponentCommonSpecUpdate(client, &oldSpec.Client.RuntimeComponentCommonSpec, &newSpec.Client.RuntimeComponentCommonSpec, false)...)
	errs = append(errs, requireUnchanged(client.Child("tieredStore"), oldSpec.Client.TieredStore, newSpec.Client.TieredStore)...)
	errs = append(errs, requireUnchanged(client.Child("cleanPolicy"), oldSpec.Client.CleanPolicy, newSpec.Client.CleanPolicy)...)

	errs = append(errs, requireUnchanged(root.Child("options"), oldSpec.Options, newSpec.Options)...)
	errs = append(errs, requireUnchanged(root.Child("podMetadata"), oldSpec.PodMetadata, newSpec.PodMetadata)...)
	errs = append(errs, requireUnchanged(root.Child("imagePullSecrets"), oldSpec.ImagePullSecrets, newSpec.ImagePullSecrets)...)
	errs = append(errs, requireUnchanged(root.Child("volumes"), oldSpec.Volumes, newSpec.Volumes)...)

	return errs
}

// validateComponentCommonSpecUpdate compares the fields shared by every
// component. syncedInPlace is true for the master and worker, whose
// runtimeVersion and resources the controller propagates to the workload; the
// client's are never propagated, so they are immutable like the rest.
func validateComponentCommonSpecUpdate(path *field.Path, oldSpec, newSpec *datav1alpha1.RuntimeComponentCommonSpec, syncedInPlace bool) field.ErrorList {
	var errs field.ErrorList

	errs = append(errs, requireUnchanged(path.Child("disabled"), oldSpec.Disabled, newSpec.Disabled)...)
	if !syncedInPlace {
		errs = append(errs, requireUnchanged(path.Child("runtimeVersion"), oldSpec.RuntimeVersion, newSpec.RuntimeVersion)...)
		errs = append(errs, requireUnchanged(path.Child("resources"), oldSpec.Resources, newSpec.Resources)...)
	}
	errs = append(errs, requireUnchanged(path.Child("options"), oldSpec.Options, newSpec.Options)...)
	errs = append(errs, requireUnchanged(path.Child("podMetadata"), oldSpec.PodMetadata, newSpec.PodMetadata)...)
	errs = append(errs, requireUnchanged(path.Child("env"), oldSpec.Env, newSpec.Env)...)
	errs = append(errs, requireUnchanged(path.Child("volumeMounts"), oldSpec.VolumeMounts, newSpec.VolumeMounts)...)
	errs = append(errs, requireUnchanged(path.Child("args"), oldSpec.Args, newSpec.Args)...)
	errs = append(errs, requireUnchanged(path.Child("nodeSelector"), oldSpec.NodeSelector, newSpec.NodeSelector)...)
	errs = append(errs, requireUnchanged(path.Child("tolerations"), oldSpec.Tolerations, newSpec.Tolerations)...)

	return errs
}

// requireUnchanged reports path as immutable when oldValue and newValue differ.
// The comparison is semantic: a nil and an empty list are the same field value
// after a round trip through the API server, and 2Gi equals 2048Mi.
func requireUnchanged(path *field.Path, oldValue, newValue interface{}) field.ErrorList {
	if equality.Semantic.DeepEqual(oldValue, newValue) {
		return nil
	}
	return field.ErrorList{field.Forbidden(path, apivalidation.FieldImmutableErrorMsg)}
}
