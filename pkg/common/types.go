/*

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

package common

import (
	corev1 "k8s.io/api/core/v1"
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
	Requests ResourceList `yaml:"requests,omitempty"`
	Limits   ResourceList `yaml:"limits,omitempty"`
}

const (
	FluidFuseBalloonKey = "fluid_fuse_balloon"

	FluidBalloonValue = "true"
)

// UserInfo to run a Container
type UserInfo struct {
	User    int `yaml:"user"`
	Group   int `yaml:"group"`
	FSGroup int `yaml:"fsGroup"`
}

// ImageInfo to run a Container
type ImageInfo struct {
	// Image of a Container
	Image string `yaml:"image"`
	// ImageTag of a Container
	ImageTag string `yaml:"imageTag"`
	// ImagePullPolicy is one of the three policies: `Always`,  `IfNotPresent`, `Never`
	ImagePullPolicy string `yaml:"imagePullPolicy"`
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
