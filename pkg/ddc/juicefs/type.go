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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

// JuiceFS The value yaml file
type JuiceFS struct {
	FullnameOverride string `yaml:"fullnameOverride"`
	Edition          string `yaml:"edition,omitempty"`
	Source           string `yaml:"source,omitempty"`

	common.ImageInfo `yaml:",inline"`
	common.UserInfo  `yaml:",inline"`

	NodeSelector map[string]string   `yaml:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	Configs      Configs             `yaml:"configs,omitempty"`
	Fuse         Fuse                `yaml:"fuse,omitempty"`
	Worker       Worker              `yaml:"worker,omitempty"`

	CacheDirs       map[string]cache       `yaml:"cacheDirs,omitempty"`
	PlacementMode   string                 `yaml:"placement,omitempty"`
	Owner           *common.OwnerReference `yaml:"owner,omitempty"`
	RuntimeIdentity common.RuntimeIdentity `yaml:"runtimeIdentity,omitempty"`
}

type Configs struct {
	Name            string `yaml:"name"`
	AccessKeySecret string `yaml:"accesskeySecret,omitempty"`
	SecretKeySecret string `yaml:"secretkeySecret,omitempty"`
	Bucket          string `yaml:"bucket,omitempty"`
	MetaUrlSecret   string `yaml:"metaurlSecret,omitempty"`
	TokenSecret     string `yaml:"tokenSecret,omitempty"`
	Storage         string `yaml:"storage,omitempty"`
	FormatCmd       string `yaml:"formatCmd,omitempty"`
}

type Worker struct {
	NotPrivileged   bool                   `yaml:"notPrivileged,omitempty"`
	Image           string                 `yaml:"image,omitempty"`
	NodeSelector    map[string]string      `yaml:"nodeSelector,omitempty"`
	ImageTag        string                 `yaml:"imageTag,omitempty"`
	ImagePullPolicy string                 `yaml:"imagePullPolicy,omitempty"`
	Resources       common.Resources       `yaml:"resources,omitempty"`
	Envs            []corev1.EnvVar        `yaml:"envs,omitempty"`
	Ports           []corev1.ContainerPort `yaml:"ports,omitempty"`
	VolumeMounts    []corev1.VolumeMount   `json:"volumeMounts,omitempty"`
	Volumes         []corev1.Volume        `json:"volumes,omitempty"`

	MountPath   string            `yaml:"mountPath,omitempty"`
	StatCmd     string            `yaml:"statCmd,omitempty"`
	Command     string            `yaml:"command,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type Fuse struct {
	NotPrivileged   bool                 `yaml:"notPrivileged,omitempty"`
	Enabled         bool                 `yaml:"enabled,omitempty"`
	Image           string               `yaml:"image,omitempty"`
	NodeSelector    map[string]string    `yaml:"nodeSelector,omitempty"`
	Envs            []corev1.EnvVar      `yaml:"envs,omitempty"`
	ImageTag        string               `yaml:"imageTag,omitempty"`
	ImagePullPolicy string               `yaml:"imagePullPolicy,omitempty"`
	Resources       common.Resources     `yaml:"resources,omitempty"`
	CriticalPod     bool                 `yaml:"criticalPod,omitempty"`
	VolumeMounts    []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes         []corev1.Volume      `json:"volumes,omitempty"`

	SubPath       string            `yaml:"subPath,omitempty"`
	MountPath     string            `yaml:"mountPath,omitempty"`
	HostMountPath string            `yaml:"hostMountPath,omitempty"`
	Command       string            `yaml:"command,omitempty"`
	StatCmd       string            `yaml:"statCmd,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	Annotations   map[string]string `yaml:"annotations,omitempty"`
}

type cache struct {
	Path string `yaml:"path,omitempty"`
	Type string `yaml:"type,omitempty"`
}

type cacheStates struct {
	cacheCapacity        string
	cached               string
	cachedPercentage     string
	cacheHitRatio        string
	cacheThroughputRatio string
}

type fuseMetrics struct {
	blockCacheBytes     int64
	blockCacheHits      int64
	blockCacheMiss      int64
	blockCacheHitsBytes int64
	blockCacheMissBytes int64
	usedSpace           int64
}
