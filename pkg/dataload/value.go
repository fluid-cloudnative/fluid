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

package dataload

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

// DataLoadValue defines the value yaml file used in DataLoad helm chart
type DataLoadValue struct {
	Name         string                 `json:"name"`
	Owner        *common.OwnerReference `json:"owner,omitempty"`
	DataLoadInfo DataLoadInfo           `json:"dataloader"`
}

// DataLoadInfo defines values used in DataLoad helm chart
type DataLoadInfo struct {
	// Policy including None, Once, Cron, OnEvent
	Policy string `json:"policy"`

	// Schedule The schedule in Cron format, only set when policy is cron, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule,omitempty"`

	// BackoffLimit specifies the upper limit times when the DataLoad job fails
	BackoffLimit int32 `json:"backoffLimit,omitempty"`

	// TargetDataset specifies the dataset that the DataLoad targets
	TargetDataset string `json:"targetDataset,omitempty"`

	// LoadMetadata specifies if the DataLoad job should load metadata from UFS when doing data load
	LoadMetadata bool `json:"loadMetadata,omitempty"`

	// TargetPaths specifies which paths should the DataLoad load
	TargetPaths []TargetPath `json:"targetPaths,omitempty"`

	// Image specifies the image that the DataLoad job uses
	Image string `json:"image,omitempty"`

	// Options specifies the extra dataload properties for runtime
	Options map[string]string `json:"options,omitempty"`

	// Labels defines labels in DataLoad's pod metadata
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations defines annotations in DataLoad's pod metadata
	Annotations map[string]string `json:"annotations,omitempty"`

	// image pull secrets
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// pod affinity
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// pod tolerations
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// node selector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// scheduler name
	SchedulerName string `json:"schedulerName,omitempty"`
}

type TargetPath struct {
	// Path specifies the path should be loaded
	Path string `json:"path,omitempty"`

	// Replicas specifies how many replicas should be loaded
	Replicas int32 `json:"replicas,omitempty"`

	// FluidNative specifies if the path is a native mountPoint(e.g. hostpath or pvc)
	FluidNative bool `json:"fluidNative,omitempty"`
}
