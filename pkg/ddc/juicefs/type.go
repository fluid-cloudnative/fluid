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

type Configs struct {
	Name               string `json:"name"`
	AccessKeySecret    string `json:"accesskeySecret,omitempty"`
	AccessKeySecretKey string `json:"accesskeySecretKey,omitempty"`
	SecretKeySecret    string `json:"secretkeySecret,omitempty"`
	SecretKeySecretKey string `json:"secretkeySecretKey,omitempty"`
	Bucket             string `json:"bucket,omitempty"`
	MetaUrlSecret      string `json:"metaurlSecret,omitempty"`
	MetaUrlSecretKey   string `json:"metaurlSecretKey,omitempty"`
	TokenSecret        string `json:"tokenSecret,omitempty"`
	TokenSecretKey     string `json:"tokenSecretKey,omitempty"`
	Storage            string `json:"storage,omitempty"`
	FormatCmd          string `json:"formatCmd,omitempty"`
	QuotaCmd           string `json:"quotaCmd,omitempty"`
}

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
