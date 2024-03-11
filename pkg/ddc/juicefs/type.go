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

package juicefs

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// JuiceFS The value json file
type JuiceFS struct {
	FullnameOverride string `json:"fullnameOverride"`
	Edition          string `json:"edition,omitempty"`
	Source           string `json:"source,omitempty"`

	common.ImageInfo `json:",inline"`
	common.UserInfo  `json:",inline"`

	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	Configs      Configs             `json:"configs,omitempty"`
	Fuse         Fuse                `json:"fuse,omitempty"`
	Worker       Worker              `json:"worker,omitempty"`

	CacheDirs       map[string]cache       `json:"cacheDirs,omitempty"`
	PlacementMode   string                 `json:"placement,omitempty"`
	Owner           *common.OwnerReference `json:"owner,omitempty"`
	RuntimeIdentity common.RuntimeIdentity `json:"runtimeIdentity,omitempty"`
}

// Configs struct describes the config of JuiceFS about the storage, including name, accessKey, bucket, meta url, etc.
type Configs struct {
	Name               string             `json:"name"`
	AccessKeySecret    string             `json:"accesskeySecret,omitempty"`
	AccessKeySecretKey string             `json:"accesskeySecretKey,omitempty"`
	SecretKeySecret    string             `json:"secretkeySecret,omitempty"`
	SecretKeySecretKey string             `json:"secretkeySecretKey,omitempty"`
	Bucket             string             `json:"bucket,omitempty"`
	MetaUrlSecret      string             `json:"metaurlSecret,omitempty"`
	MetaUrlSecretKey   string             `json:"metaurlSecretKey,omitempty"`
	TokenSecret        string             `json:"tokenSecret,omitempty"`
	TokenSecretKey     string             `json:"tokenSecretKey,omitempty"`
	Storage            string             `json:"storage,omitempty"`
	FormatCmd          string             `json:"formatCmd,omitempty"`
	QuotaCmd           string             `json:"quotaCmd,omitempty"`
	EncryptEnvOptions  []EncryptEnvOption `json:"encryptEnvOptions,omitempty"`
}

type EncryptEnvOption struct {
	Name             string `json:"name"`    //  name
	EnvName          string `json:"envName"` //  envName
	SecretKeyRefName string `json:"secretKeyRefName"`
	SecretKeyRefKey  string `json:"secretKeyRefKey"`
}

// Worker struct describes the configuration of JuiceFS worker node, including image, resources, environmental variables, volume mounts, etc.
type Worker struct {
	Privileged      bool                 `json:"privileged"`
	Image           string               `json:"image,omitempty"`
	NodeSelector    map[string]string    `json:"nodeSelector,omitempty"`
	ImageTag        string               `json:"imageTag,omitempty"`
	ImagePullPolicy string               `json:"imagePullPolicy,omitempty"`
	Resources       common.Resources     `json:"resources,omitempty"`
	Envs            []corev1.EnvVar      `json:"envs,omitempty"`
	VolumeMounts    []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes         []corev1.Volume      `json:"volumes,omitempty"`
	HostNetwork     bool                 `json:"hostNetwork,omitempty"`
	MetricsPort     *int                 `json:"metricsPort,omitempty"`

	MountPath   string            `json:"mountPath,omitempty"`
	StatCmd     string            `json:"statCmd,omitempty"`
	Command     string            `json:"command,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Fuse struct holds the configuration of JuiceFS FUSE (File System in User Space).
type Fuse struct {
	Privileged      bool                 `json:"privileged"`
	Enabled         bool                 `json:"enabled,omitempty"`
	Image           string               `json:"image,omitempty"`
	NodeSelector    map[string]string    `json:"nodeSelector,omitempty"`
	Envs            []corev1.EnvVar      `json:"envs,omitempty"`
	ImageTag        string               `json:"imageTag,omitempty"`
	ImagePullPolicy string               `json:"imagePullPolicy,omitempty"`
	Resources       common.Resources     `json:"resources,omitempty"`
	CriticalPod     bool                 `json:"criticalPod,omitempty"`
	VolumeMounts    []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes         []corev1.Volume      `json:"volumes,omitempty"`
	HostNetwork     bool                 `json:"hostNetwork,omitempty"`
	HostPID         bool                 `json:"hostPID,omitempty"`
	MetricsPort     *int                 `json:"metricsPort,omitempty"`

	SubPath       string            `json:"subPath,omitempty"`
	MountPath     string            `json:"mountPath,omitempty"`
	HostMountPath string            `json:"hostMountPath,omitempty"`
	Command       string            `json:"command,omitempty"`
	StatCmd       string            `json:"statCmd,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

type cache struct {
	Path         string                 `json:"path,omitempty"`
	Type         string                 `json:"type,omitempty"`
	VolumeSource *v1alpha1.VolumeSource `json:"volumeSource,omitempty"`
}

// cacheStates struct holds various cache states for a JuiceFS filesystem.
type cacheStates struct {
	cacheCapacity        string
	cached               string
	cachedPercentage     string
	cacheHitRatio        string
	cacheThroughputRatio string
}

// fuseMetrics struct holds various FUSE (File System in User Space) related metrics for a JuiceFS filesystem.
type fuseMetrics struct {
	blockCacheBytes     int64
	blockCacheHits      int64
	blockCacheMiss      int64
	blockCacheHitsBytes int64
	blockCacheMissBytes int64
	usedSpace           int64
}
