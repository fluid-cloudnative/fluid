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
package dataprocess

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type DataProcessValue struct {
	Name string `json:"name"`

	Owner *common.OwnerReference `json:"owner,omitempty"`

	DataProcessInfo DataProcessInfo `json:"dataProcess"`
}

type DataProcessInfo struct {
	TargetDataset string `json:"targetDataset,omitempty"`

	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	Annotations map[string]string `json:"annotations,omitempty"`

	JobProcessor *JobProcessor `json:"jobProcessor,omitempty"`

	ScriptProcessor *ScriptProcessor `json:"scriptProcessor,omitempty"`
}

type ScriptProcessor struct {
	Image string `json:"image,omitempty"`

	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty"`

	Command []string `json:"command,omitempty"`

	Source string `json:"source,omitempty"`

	Envs []corev1.EnvVar `json:"envs,omitempty"`

	Volumes []corev1.Volume `json:"volumes,omitempty"`

	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

type JobProcessor struct {
	PodSpec *corev1.PodSpec `json:"podSpec,omitempty"`
}
