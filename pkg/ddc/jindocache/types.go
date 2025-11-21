/*
Copyright 2023 The Fluid Author.

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

package jindocache

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type Jindo struct {
	FullnameOverride    string                        `json:"fullnameOverride"`
	OwnerDatasetId      string                        `json:"ownerDatasetId"`
	Image               string                        `json:"image"`
	ImageTag            string                        `json:"imageTag"`
	ImagePullPolicy     string                        `json:"imagePullPolicy"`
	FuseImage           string                        `json:"fuseImage"`
	FuseImageTag        string                        `json:"fuseImageTag"`
	FuseImagePullPolicy string                        `json:"fuseImagePullPolicy"`
	User                int                           `json:"user"`
	Group               int                           `json:"group"`
	UseHostNetwork      bool                          `json:"useHostNetwork"`
	Properties          map[string]string             `json:"properties"`
	Master              Master                        `json:"master"`
	Worker              Worker                        `json:"worker"`
	Fuse                Fuse                          `json:"fuse"`
	Mounts              Mounts                        `json:"mounts"`
	HadoopConfig        HadoopConfig                  `json:"hadoopConfig,omitempty"`
	Secret              string                        `json:"secret,omitempty"`
	Tolerations         []corev1.Toleration           `json:"tolerations,omitempty"`
	InitPortCheck       common.InitPortCheck          `json:"initPortCheck,omitempty"`
	LogConfig           map[string]string             `json:"logConfig,omitempty"`
	FuseLogConfig       map[string]string             `json:"fuseLogConfig,omitempty"`
	PlacementMode       string                        `json:"placement,omitempty"`
	Owner               *common.OwnerReference        `json:"owner,omitempty"`
	RuntimeIdentity     common.RuntimeIdentity        `json:"runtimeIdentity"`
	ClusterDomain       string                        `json:"clusterDomain,omitempty"`
	UFSVolumes          []UFSVolume                   `json:"ufsVolumes,omitempty"`
	SecretKey           string                        `json:"secretKey,omitempty"`
	SecretValue         string                        `json:"secretValue,omitempty"`
	UseStsToken         bool                          `json:"UseStsToken"`
	MountType           string                        `json:"mountType,omitempty"`
	ImagePullSecrets    []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

type HadoopConfig struct {
	ConfigMap       string `json:"configMap"`
	IncludeHdfsSite bool   `json:"includeHdfsSite"`
	IncludeCoreSite bool   `json:"includeCoreSite"`
}

type Master struct {
	ReplicaCount        int                           `json:"replicaCount"`
	Resources           common.Resources              `json:"resources"`
	NodeSelector        map[string]string             `json:"nodeSelector,omitempty"`
	MasterProperties    map[string]string             `json:"properties"`
	FileStoreProperties map[string]string             `json:"fileStoreProperties"`
	TokenProperties     map[string]string             `json:"secretProperties"`
	Port                Ports                         `json:"ports,omitempty"`
	OssKey              string                        `json:"osskey,omitempty"`
	OssSecret           string                        `json:"osssecret,omitempty"`
	Tolerations         []corev1.Toleration           `json:"tolerations,omitempty"`
	DnsServer           string                        `json:"dnsServer,omitempty"`
	NameSpace           string                        `json:"namespace,omitempty"`
	Labels              map[string]string             `json:"labels,omitempty"`
	Annotations         map[string]string             `json:"annotations,omitempty"`
	ServiceCount        int                           `json:"svccount"`
	Env                 map[string]string             `json:"env,omitempty"`
	CacheSets           map[string]*CacheSet          `json:"cachesets,omitempty"`
	VolumeMounts        []corev1.VolumeMount          `json:"volumeMounts,omitempty"`
	Volumes             []corev1.Volume               `json:"volumes,omitempty"`
	ImagePullSecrets    []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

type Worker struct {
	ReplicaCount     int                 `json:"replicaCount"`
	Resources        common.Resources    `json:"resources,omitempty"`
	NodeSelector     map[string]string   `json:"nodeSelector,omitempty"`
	WorkerProperties map[string]string   `json:"properties"`
	Port             Ports               `json:"ports,omitempty"`
	Tolerations      []corev1.Toleration `json:"tolerations,omitempty"`
	// Affinity         corev1.Affinity       `json:"affinity,omitempty"`
	Labels           map[string]string             `json:"labels,omitempty"`
	Annotations      map[string]string             `json:"annotations,omitempty"`
	Path             string                        `json:"dataPath"`
	Env              map[string]string             `json:"env,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

type Ports struct {
	Rpc  int `json:"rpc,omitempty"`
	Raft int `json:"raft,omitempty"`
}

type Fuse struct {
	Args             []string                      `json:"args"`
	HostPath         string                        `json:"hostPath"`
	NodeSelector     map[string]string             `json:"nodeSelector,omitempty"`
	FuseProperties   map[string]string             `json:"properties"`
	Global           bool                          `json:"global,omitempty"`
	RunAs            string                        `json:"runAs,omitempty"`
	Tolerations      []corev1.Toleration           `json:"tolerations,omitempty"`
	Labels           map[string]string             `json:"labels,omitempty"`
	Annotations      map[string]string             `json:"annotations,omitempty"`
	CriticalPod      bool                          `json:"criticalPod,omitempty"`
	Resources        common.Resources              `json:"resources,omitempty"`
	MountPath        string                        `json:"mountPath,omitempty"`
	Mode             string                        `json:"mode,omitempty"`
	Env              map[string]string             `json:"env,omitempty"`
	HostPID          bool                          `json:"hostPID,omitempty"`
	MetricsPort      int                           `json:"metricsPort,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

type Mounts struct {
	Master            map[string]*Level `json:"master"`
	WorkersAndClients map[string]*Level `json:"workersAndClients"`
}

type Resource struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Level struct {
	Path       string `json:"path,omitempty"`
	Type       string `json:"type,omitempty"`
	MediumType string `json:"mediumType,omitempty"`
	Quota      string `json:"quota,omitempty"`
}

type cacheStates struct {
	cacheCapacity string
	// cacheable        string
	// lowWaterMark     string
	// highWaterMark    string
	cached           string
	cachedPercentage string
	// nonCacheable     string
}

type UFSVolume struct {
	Name          string `json:"name"`
	SubPath       string `json:"subPath,omitempty"`
	ContainerPath string `json:"containerPath"`
	ReadOnly      bool   `json:"readOnly"`
}

type CacheSet struct {
	Name              string `json:"name"`
	Path              string `json:"path,omitempty"`
	CacheStrategy     string `json:"cacheStrategy"`
	MetaPolicy        string `json:"metaPolicy"`
	ReadPolicy        string `json:"readPolicy"`
	WritePolicy       string `json:"writePolicy"`
	ReadCacheReplica  int    `json:"readCacheReplica"`
	WriteCacheReplica int    `json:"writeCacheReplica"`
}
