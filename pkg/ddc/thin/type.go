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

package thin

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type ThinValue struct {
	FullnameOverride string `json:"fullnameOverride"`

	common.ImageInfo `json:",inline"`
	common.UserInfo  `json:",inline"`

	Fuse            Fuse                   `json:"fuse,omitempty"`
	Worker          Worker                 `json:"worker,omitempty"`
	NodeSelector    map[string]string      `json:"nodeSelector,omitempty"`
	Tolerations     []corev1.Toleration    `json:"tolerations,omitempty"`
	PlacementMode   string                 `json:"placement,omitempty"`
	Owner           *common.OwnerReference `json:"owner,omitempty"`
	RuntimeValue    string                 `json:"runtimeValue"`
	RuntimeIdentity common.RuntimeIdentity `json:"runtimeIdentity"`
}

type Worker struct {
	Image           string                 `json:"image,omitempty"`
	ImageTag        string                 `json:"imageTag,omitempty"`
	ImagePullPolicy string                 `json:"imagePullPolicy,omitempty"`
	Resources       common.Resources       `json:"resources,omitempty"`
	NodeSelector    map[string]string      `json:"nodeSelector,omitempty"`
	HostNetwork     bool                   `json:"hostNetwork,omitempty"`
	Envs            []corev1.EnvVar        `json:"envs,omitempty"`
	Ports           []corev1.ContainerPort `json:"ports,omitempty"`
	Volumes         []corev1.Volume        `json:"volumes,omitempty"`
	VolumeMounts    []corev1.VolumeMount   `json:"volumeMounts,omitempty"`
	LivenessProbe   *corev1.Probe          `json:"livenessProbe,omitempty"`
	ReadinessProbe  *corev1.Probe          `json:"readinessProbe,omitempty"`
	CacheDir        string                 `json:"cacheDir,omitempty"`
}

type Fuse struct {
	Enabled         bool                   `json:"enabled,omitempty"`
	Image           string                 `json:"image,omitempty"`
	ImageTag        string                 `json:"imageTag,omitempty"`
	ImagePullPolicy string                 `json:"imagePullPolicy,omitempty"`
	Resources       common.Resources       `json:"resources,omitempty"`
	Ports           []corev1.ContainerPort `json:"ports,omitempty"`
	CriticalPod     bool                   `json:"criticalPod,omitempty"`
	HostNetwork     bool                   `json:"hostNetwork,omitempty"`
	TargetPath      string                 `json:"targetPath,omitempty"`
	NodeSelector    map[string]string      `json:"nodeSelector,omitempty"`
	Envs            []corev1.EnvVar        `json:"envs,omitempty"`
	Command         []string               `json:"command,omitempty"`
	Args            []string               `json:"args,omitempty"`
	Volumes         []corev1.Volume        `json:"volumes,omitempty"`
	VolumeMounts    []corev1.VolumeMount   `json:"volumeMounts,omitempty"`
	LivenessProbe   *corev1.Probe          `json:"livenessProbe,omitempty"`
	ReadinessProbe  *corev1.Probe          `json:"readinessProbe,omitempty"`
	CacheDir        string                 `json:"cacheDir,omitempty"`
	ConfigValue     string                 `json:"configValue"`
	ConfigStorage   string                 `json:"configStorage"`
}

type Config struct {
	Mounts                       []datav1alpha1.Mount                         `json:"mounts,omitempty"`
	TargetPath                   string                                       `json:"targetPath,omitempty"`
	RuntimeOptions               map[string]string                            `json:"runtimeOptions,omitempty"`
	PersistentVolumeAttrs        map[string]*corev1.CSIPersistentVolumeSource `json:"persistentVolumeAttrs,omitempty"`
	PersistentVolumeMountOptions map[string][]string                          `json:"persistentVolumeMountOptions,omitempty"`
}

// RuntimeSetConfig is with the info of the workers and fuses
type RuntimeSetConfig struct {
	Workers []string `json:"workers"`
	Fuses   []string `json:"fuses"`
}
