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

package efc

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

// The value yaml file
type EFC struct {
	FullnameOverride string                 `yaml:"fullnameOverride"`
	PlacementMode    string                 `yaml:"placement,omitempty"`
	Master           Master                 `yaml:"master"`
	Worker           Worker                 `yaml:"worker"`
	Fuse             Fuse                   `yaml:"fuse"`
	InitFuse         InitFuse               `yaml:"initFuse"`
	OSAdvise         OSAdvise               `yaml:"osAdvise"`
	Tolerations      []v1.Toleration        `yaml:"tolerations,omitempty"`
	Owner            *common.OwnerReference `yaml:"owner,omitempty"`
	RuntimeIdentity  common.RuntimeIdentity `yaml:"runtimeIdentity,omitempty"`
}

type OSAdvise struct {
	OSVersion string `yaml:"osVersion,omitempty"`
	Enabled   bool   `yaml:"enabled"`
}

type Master struct {
	common.ImageInfo `yaml:",inline"`
	MountPoint       string            `yaml:"mountPoint,omitempty"`
	Replicas         int32             `yaml:"count,omitempty"`
	Enabled          bool              `yaml:"enabled"`
	Options          string            `yaml:"option,omitempty"`
	Resources        common.Resources  `yaml:"resources,omitempty"`
	NodeSelector     map[string]string `yaml:"nodeSelector,omitempty"`
	HostNetwork      bool              `yaml:"hostNetwork"`
	TieredStore      TieredStore       `yaml:"tieredstore,omitempty"`
	Labels           map[string]string `yaml:"labels,omitempty"`
	Annotations      map[string]string `yaml:"annotations,omitempty"`
}

type Worker struct {
	common.ImageInfo `yaml:",inline"`
	Port             Port              `yaml:"port,omitempty"`
	Enabled          bool              `yaml:"enabled"`
	Options          string            `yaml:"option,omitempty"`
	Resources        common.Resources  `yaml:"resources,omitempty"`
	NodeSelector     map[string]string `yaml:"nodeSelector,omitempty"`
	HostNetwork      bool              `yaml:"hostNetwork"`
	TieredStore      TieredStore       `yaml:"tieredstore,omitempty"`
	Labels           map[string]string `yaml:"labels,omitempty"`
	Annotations      map[string]string `yaml:"annotations,omitempty"`
}

type Fuse struct {
	common.ImageInfo `yaml:",inline"`
	MountPoint       string            `yaml:"mountPoint,omitempty"`
	HostMountPath    string            `yaml:"hostMountPath,omitempty"`
	Port             Port              `yaml:"port,omitempty"`
	Options          string            `yaml:"option,omitempty"`
	Resources        common.Resources  `yaml:"resources,omitempty"`
	NodeSelector     map[string]string `yaml:"nodeSelector,omitempty"`
	HostNetwork      bool              `yaml:"hostNetwork"`
	TieredStore      TieredStore       `yaml:"tieredstore,omitempty"`
	CriticalPod      bool              `yaml:"criticalPod"`
	Labels           map[string]string `yaml:"labels,omitempty"`
	Annotations      map[string]string `yaml:"annotations,omitempty"`
}

type InitFuse struct {
	common.ImageInfo `yaml:",inline"`
}

type TieredStore struct {
	Levels []Level `yaml:"levels,omitempty"`
}

type Level struct {
	Alias      string `yaml:"alias,omitempty"`
	Level      int    `yaml:"level"`
	MediumType string `yaml:"mediumtype,omitempty"`
	Type       string `yaml:"type,omitempty"`
	Path       string `yaml:"path,omitempty"`
	Quota      string `yaml:"quota,omitempty"`
	High       string `yaml:"high,omitempty"`
	Low        string `yaml:"low,omitempty"`
}

type Port struct {
	Rpc     int `yaml:"rpc,omitempty"`
	Monitor int `yaml:"monitor,omitempty"`
}

type WorkerEndPoints struct {
	ContainerEndPoints []string `json:"containerendpoints,omitempty"`
}

type cacheHitStates struct {
	cacheHitRatio  string
	localHitRatio  string
	remoteHitRatio string

	localThroughputRatio  string
	remoteThroughputRatio string
	cacheThroughputRatio  string

	//bytesReadLocal int64
	//bytesReadRemote int64
	//bytesReadUfsAll int64

	//timestamp time.Time
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

type MountInfo struct {
	MountPoint       string
	MountPointPrefix string
	ServiceAddr      string
	FileSystemId     string
	DirPath          string
}

func (value *EFC) getTiredStoreLevel0Path() (path string) {
	for _, level := range value.Worker.TieredStore.Levels {
		if level.Level == 0 {
			path = level.Path
			break
		}
	}
	return
}

func (value *EFC) getTiredStoreLevel0Type() (t string) {
	for _, level := range value.Worker.TieredStore.Levels {
		if level.Level == 0 {
			t = level.Type
			break
		}
	}
	return
}

func (value *EFC) getTiredStoreLevel0MediumType() (t string) {
	for _, level := range value.Worker.TieredStore.Levels {
		if level.Level == 0 {
			t = level.MediumType
			break
		}
	}
	return
}

func (value *EFC) getTiredStoreLevel0Quota() (quota string) {
	for _, level := range value.Worker.TieredStore.Levels {
		if level.Level == 0 {
			quota = level.Quota
			break
		}
	}
	return
}

func (e *EFCEngine) getDefaultTiredStoreLevel0() Level {
	return Level{
		Level:      0,
		Type:       string(common.VolumeTypeEmptyDir),
		Path:       fmt.Sprintf("%s/%s/%s", "/cache_dir", e.namespace, e.name),
		MediumType: string(common.Memory),
		Quota:      utils.TransformQuantityToEFCUnit(&miniWorkerQuota),
	}
}
