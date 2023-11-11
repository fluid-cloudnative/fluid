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

package alluxio

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// The value yaml file
type Alluxio struct {
	FullnameOverride string `json:"fullnameOverride"`

	common.ImageInfo `json:",inline"`
	common.UserInfo  `json:",inline"`

	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	JvmOptions   []string          `json:"jvmOptions,omitempty"`

	// Note: the mount secret options may be in here if existing direct mount root path in alluxio
	// So, alluxio master pod should handle with replacing the secret file path.
	Properties map[string]string `json:"properties,omitempty"`

	Master Master `json:"master,omitempty"`

	JobMaster JobMaster `json:"jobMaster,omitempty"`

	Worker Worker `json:"worker,omitempty"`

	JobWorker JobWorker `json:"jobWorker,omitempty"`

	Fuse Fuse `json:"fuse,omitempty"`

	APIGateway APIGateway `json:"apiGateway,omitempty"`

	TieredStore TieredStore `json:"tieredstore,omitempty"`

	Metastore Metastore `json:"metastore,omitempty"`

	Journal Journal `json:"journal,omitempty"`

	ShortCircuit ShortCircuit `json:"shortCircuit,omitempty"`
	// Enablefluid bool `json:"enablefluid,omitempty"`

	UFSPaths []UFSPath `json:"ufsPaths,omitempty"`

	UFSVolumes []UFSVolume `json:"ufsVolumes,omitempty"`

	InitUsers common.InitUsers `json:"initUsers,omitempty"`

	Monitoring string `json:"monitoring,omitempty"`

	HadoopConfig HadoopConfig `json:"hadoopConfig,omitempty"`

	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	PlacementMode string `json:"placement,omitempty"`

	RuntimeIdentity common.RuntimeIdentity `json:"runtimeIdentity"`

	Owner *common.OwnerReference `json:"owner,omitempty"`
}

type HadoopConfig struct {
	ConfigMap       string `json:"configMap"`
	IncludeHdfsSite bool   `json:"includeHdfsSite"`
	IncludeCoreSite bool   `json:"includeCoreSite"`
}

type UFSPath struct {
	HostPath  string `json:"hostPath"`
	UFSVolume `json:",inline"`
}

type UFSVolume struct {
	Name          string `json:"name"`
	SubPath       string `json:"subPath,omitempty"`
	ContainerPath string `json:"containerPath"`
}

type Metastore struct {
	VolumeType string `json:"volumeType,omitempty"`
	Size       string `json:"size,omitempty"`
}

type Journal struct {
	VolumeType string `json:"volumeType,omitempty"`
	Size       string `json:"size,omitempty"`
}

type ShortCircuit struct {
	Enable     bool   `json:"enable,omitempty"`
	Policy     string `json:"policy,omitempty"`
	VolumeType string `json:"volumeType,omitempty"`
}

type Ports struct {
	Rpc      int `json:"rpc,omitempty"`
	Web      int `json:"web,omitempty"`
	Embedded int `json:"embedded,omitempty"`
	Data     int `json:"data,omitempty"`
	Rest     int `json:"rest,omitempty"`
}

type APIGateway struct {
	Enabled bool  `json:"enabled,omitempty"`
	Ports   Ports `json:"ports,omitempty"`
}

type JobMaster struct {
	Ports     Ports            `json:"ports,omitempty"`
	Resources common.Resources `json:"resources,omitempty"`
}

type JobWorker struct {
	Ports     Ports            `json:"ports,omitempty"`
	Resources common.Resources `json:"resources,omitempty"`
}

type Worker struct {
	JvmOptions   []string             `json:"jvmOptions,omitempty"`
	Env          map[string]string    `json:"env,omitempty"`
	NodeSelector map[string]string    `json:"nodeSelector,omitempty"`
	Properties   map[string]string    `json:"properties,omitempty"`
	HostNetwork  bool                 `json:"hostNetwork,omitempty"`
	Resources    common.Resources     `json:"resources,omitempty"`
	Ports        Ports                `json:"ports,omitempty"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes      []corev1.Volume      `json:"volumes,omitempty"`
	Labels       map[string]string    `json:"labels,omitempty"`
	Annotations  map[string]string    `json:"annotations,omitempty"`
}

type Master struct {
	JvmOptions   []string             `json:"jvmOptions,omitempty"`
	Env          map[string]string    `json:"env,omitempty"`
	Affinity     Affinity             `json:"affinity"`
	NodeSelector map[string]string    `json:"nodeSelector,omitempty"`
	Properties   map[string]string    `json:"properties,omitempty"`
	Replicas     int32                `json:"replicaCount,omitempty"`
	HostNetwork  bool                 `json:"hostNetwork,omitempty"`
	Resources    common.Resources     `json:"resources,omitempty"`
	Ports        Ports                `json:"ports,omitempty"`
	BackupPath   string               `json:"backupPath,omitempty"`
	Restore      Restore              `json:"restore,omitempty"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes      []corev1.Volume      `json:"volumes,omitempty"`
	Labels       map[string]string    `json:"labels,omitempty"`
	Annotations  map[string]string    `json:"annotations,omitempty"`
	// config storage to store mount config.
	MountConfigStorage string `json:"mountConfigStorage,omitempty"`
	// non native mount infos when using configmap as mount storage.
	NonNativeMounts []string `json:"nonNativeMounts,omitempty"`
}

type Restore struct {
	Enabled bool   `json:"enabled,omitempty"`
	Path    string `json:"path,omitempty"`
	PVCName string `json:"pvcName,omitempty"`
}

type Fuse struct {
	Image              string               `json:"image,omitempty"`
	NodeSelector       map[string]string    `json:"nodeSelector,omitempty"`
	ImageTag           string               `json:"imageTag,omitempty"`
	ImagePullPolicy    string               `json:"imagePullPolicy,omitempty"`
	Properties         map[string]string    `json:"properties,omitempty"`
	Env                map[string]string    `json:"env,omitempty"`
	JvmOptions         []string             `json:"jvmOptions,omitempty"`
	MountPath          string               `json:"mountPath,omitempty"`
	ShortCircuitPolicy string               `json:"shortCircuitPolicy,omitempty"`
	Args               []string             `json:"args,omitempty"`
	HostNetwork        bool                 `json:"hostNetwork,omitempty"`
	Enabled            bool                 `json:"enabled,omitempty"`
	Resources          common.Resources     `json:"resources,omitempty"`
	Global             bool                 `json:"global,omitempty"`
	CriticalPod        bool                 `json:"criticalPod,omitempty"`
	VolumeMounts       []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes            []corev1.Volume      `json:"volumes,omitempty"`
	Labels             map[string]string    `json:"labels,omitempty"`
	Annotations        map[string]string    `json:"annotations,omitempty"`
}

type TieredStore struct {
	Levels []Level `json:"levels,omitempty"`
}

type Level struct {
	Alias      string `json:"alias,omitempty"`
	Level      int    `json:"level"`
	MediumType string `json:"mediumtype,omitempty"`
	Type       string `json:"type,omitempty"`
	Path       string `json:"path,omitempty"`
	Quota      string `json:"quota,omitempty"`
	High       string `json:"high,omitempty"`
	Low        string `json:"low,omitempty"`
}

type Affinity struct {
	NodeAffinity *NodeAffinity `json:"nodeAffinity"`
}

type cacheHitStates struct {
	cacheHitRatio  string
	localHitRatio  string
	remoteHitRatio string

	localThroughputRatio  string
	remoteThroughputRatio string
	cacheThroughputRatio  string

	bytesReadLocal  int64
	bytesReadRemote int64
	bytesReadUfsAll int64

	timestamp time.Time
}

type cacheStates struct {
	cacheCapacity string
	// cacheable        string
	// lowWaterMark     string
	// highWaterMark    string
	cached           string
	cachedPercentage string
	cacheHitStates   cacheHitStates
	// nonCacheable     string
}

func (value *Alluxio) getTiredStoreLevel0Path(name, namespace string) (path string) {
	path = fmt.Sprintf("/dev/shm/%s/%s", namespace, name)
	for _, level := range value.TieredStore.Levels {
		if level.Level == 0 {
			path = level.Path
			break
		}
	}
	return
}
