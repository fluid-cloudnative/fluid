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
package datamigrate

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

type DataMigrateValue struct {
	Name            string                 `json:"name"`
	Owner           *common.OwnerReference `json:"owner,omitempty"`
	DataMigrateInfo DataMigrateInfo        `json:"datamigrate"`
}

type DataMigrateInfo struct {
	// Policy for migrate, including None, Once, Cron, OnEvent
	Policy string `json:"policy"`

	// Schedule The schedule in Cron format, only set when policy is cron, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule,omitempty"`

	// BackoffLimit specifies the upper limit times when the DataMigrate job fails
	BackoffLimit int32 `json:"backoffLimit,omitempty"`

	// TargetDataset specifies the dataset that the DataLoad targets
	TargetDataset string `json:"targetDataset,omitempty"`

	// MigrateFrom specifies the data that the DataMigrate migrate from
	MigrateFrom string `json:"migrateFrom,omitempty"`

	// MigrateTo specifies the data that the DataMigrate migrate to
	MigrateTo string `json:"migrateTo,omitempty"`

	// EncryptOptions specifies the encrypt options that the DataMigrate job uses
	EncryptOptions []v1alpha1.EncryptOption `json:"encryptOptions,omitempty"`

	// Image specifies the image that the DataMigrate job uses
	Image string `json:"image,omitempty"`

	// Options specifies the extra dataMigrate properties for runtime
	Options map[string]string `json:"options,omitempty"`

	// Labels defines labels in DataMigrate's pod metadata
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations defines annotations in DataMigrate's pod metadata
	Annotations map[string]string `json:"annotations,omitempty"`

	// image pull secrets
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// specifies local:// and pvc:// volume
	NativeVolumes []corev1.Volume `json:"nativeVolumes,omitempty"`

	// specifies local:// and pvc:// volume mount
	NativeVolumeMounts []corev1.VolumeMount `json:"nativeVolumeMounts,omitempty"`

	// specifies pod affinity
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// specifies node selector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// specifies pod tolerations
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// specifies scheduler name
	SchedulerName string `json:"schedulerName,omitempty"`
}
