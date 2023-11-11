/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package common

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type RuntimeRole string

// CacheStateName is the name identifying various cacheStateName in a CacheStateNameList.
type CacheStateName string

// ResourceList is a set of (resource name, quantity) pairs.
type CacheStateList map[CacheStateName]string

// CacheStateName names must be not more than 63 characters, consisting of upper- or lower-case alphanumeric characters,
// with the -, _, and . characters allowed anywhere, except the first or last character.
// The default convention, matching that for annotations, is to use lower-case names, with dashes, rather than
// camel case, separating compound words.
// Fully-qualified resource typenames are constructed from a DNS-style subdomain, followed by a slash `/` and a name.
const (
	// Cached in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	Cached CacheStateName = "cached"
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	// Cacheable CacheStateName = "cacheable"
	LowWaterMark CacheStateName = "lowWaterMark"
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	HighWaterMark CacheStateName = "highWaterMark"
	// NonCacheable size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)
	NonCacheable CacheStateName = "nonCacheable"
	// Percentage represents the cache percentage over the total data in the underlayer filesystem.
	// 1.5 = 1500m
	CachedPercentage CacheStateName = "cachedPercentage"

	CacheCapacity CacheStateName = "cacheCapacity"

	// CacheHitRatio defines total cache hit ratio(both local hit and remote hit), it is a metric to learn
	// how much profit a distributed cache brings.
	CacheHitRatio CacheStateName = "cacheHitRatio"

	// LocalHitRatio defines local hit ratio. It represents how many data is requested from local cache worker
	LocalHitRatio CacheStateName = "localHitRatio"

	// RemoteHitRatio defines remote hit ratio. It represents how many data is requested from remote cache worker(s).
	RemoteHitRatio CacheStateName = "remoteHitRatio"

	// CacheThroughputRatio defines total cache hit throughput ratio, both local hit and remote hit are included.
	CacheThroughputRatio CacheStateName = "cacheThroughputRatio"

	// LocalThroughputRatio defines local cache hit throughput ratio.
	LocalThroughputRatio CacheStateName = "localThroughputRatio"

	// RemoteThroughputRatio defines remote cache hit throughput ratio.
	RemoteThroughputRatio CacheStateName = "remoteThroughputRatio"
)

type ResourceList map[corev1.ResourceName]string

type Resources struct {
	Requests ResourceList `json:"requests,omitempty" yaml:"requests,omitempty"`
	Limits   ResourceList `json:"limits,omitempty" yaml:"limits,omitempty"`
}

const (
	FluidFuseBalloonKey = "fluid_fuse_balloon"

	FluidBalloonValue = "true"
)

// UserInfo to run a Container
type UserInfo struct {
	User    int `json:"user" yaml:"user"`
	Group   int `json:"group" yaml:"group"`
	FSGroup int `json:"fsGroup" yaml:"fsGroup"`
}

// ImageInfo to run a Container
type ImageInfo struct {
	// Image of a Container
	Image string `json:"image" yaml:"image"`
	// ImageTag of a Container
	ImageTag string `json:"imageTag" yaml:"imageTag"`
	// ImagePullPolicy is one of the three policies: `Always`,  `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy" yaml:"imagePullPolicy"`
	// ImagePullSecrets
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets" yaml:"imagePullSecrets"`
}

// Phase is a valid value of a task stage
type Phase string

// These are possible phases of a Task
const (
	PhaseNone      Phase = ""
	PhasePending   Phase = "Pending"
	PhaseExecuting Phase = "Executing"
	PhaseComplete  Phase = "Complete"
	PhaseFailed    Phase = "Failed"
)

// ConditionType is a valid value for Condition.Type
type ConditionType string

// These are valid conditions of a Task
const (
	// Complete means the task has completed its execution.
	Complete ConditionType = "Complete"
	// Failed means the task has failed its execution.
	Failed ConditionType = "Failed"
)

type OwnerReference struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
	// API version of the referent.
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	// Kind of the referent.
	Kind string `json:"kind" yaml:"kind"`
	// Name of the referent.
	Name string `json:"name" yaml:"name"`
	// UID of the referent.
	UID string `json:"uid" yaml:"uid"`
	// If true, this reference points to the managing controller.
	// +optional
	Controller bool `json:"controller" yaml:"controller"`
	// If true, AND if the owner has the "foregroundDeletion" finalizer, then
	// +optional
	BlockOwnerDeletion bool `json:"blockOwnerDeletion" yaml:"blockOwnerDeletion"`
}

// FuseInjectionTemplate for injecting fuse container into the pod
type FuseInjectionTemplate struct {
	PVCName              string
	SubPath              string
	FuseContainer        corev1.Container
	VolumeMountsToUpdate []corev1.VolumeMount
	VolumeMountsToAdd    []corev1.VolumeMount
	VolumesToUpdate      []corev1.Volume
	VolumesToAdd         []corev1.Volume

	FuseMountInfo FuseMountInfo
}

type FuseMountInfo struct {
	SubPath            string
	HostMountPath      string
	ContainerMountPath string
	FsType             string
}

// FuseSidecarInjectOption are options for webhook to inject fuse sidecar containers
type FuseSidecarInjectOption struct {
	EnableCacheDir             bool
	EnableUnprivilegedSidecar  bool
	SkipSidecarPostStartInject bool
}

func (f FuseSidecarInjectOption) String() string {
	return fmt.Sprintf("EnableCacheDir=%v;EnableUnprivilegedSidecar=%v",
		f.EnableCacheDir,
		f.EnableUnprivilegedSidecar)
}

// The Application which is using Fluid,
// and it has serveral PodSpecs.
type FluidApplication interface {
	GetPodSpecs() (specs []FluidObject, err error)

	SetPodSpecs(specs []FluidObject) (err error)

	// GetObject gets K8s object which can be consumed by K8s API
	GetObject() runtime.Object
}

// FluidObject simulates the V1 Pod Spec, it has v1.volumes, v1.containers inside
type FluidObject interface {
	GetRoot() runtime.Object

	GetVolumes() (volumes []corev1.Volume, err error)

	SetVolumes(volumes []corev1.Volume) (err error)

	GetInitContainers() (containers []corev1.Container, err error)

	GetContainers() (containers []corev1.Container, err error)

	SetContainers(containers []corev1.Container) (err error)

	SetInitContainers(containers []corev1.Container) (err error)

	GetVolumeMounts() (volumeMounts []corev1.VolumeMount, err error)

	SetMetaObject(metaObject metav1.ObjectMeta) (err error)

	GetMetaObject() (metaObject metav1.ObjectMeta, err error)
}
