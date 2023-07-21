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
	v1 "k8s.io/api/core/v1"
)

type Jindo struct {
	Image               string                 `yaml:"image"`
	ImageTag            string                 `yaml:"imageTag"`
	ImagePullPolicy     string                 `yaml:"imagePullPolicy"`
	FuseImage           string                 `yaml:"fuseImage"`
	FuseImageTag        string                 `yaml:"fuseImageTag"`
	FuseImagePullPolicy string                 `yaml:"fuseImagePullPolicy"`
	User                int                    `yaml:"user"`
	Group               int                    `yaml:"group"`
	FsGroup             int                    `yaml:"fsGroup"`
	UseHostNetwork      bool                   `yaml:"useHostNetwork"`
	UseHostPID          bool                   `yaml:"useHostPID"`
	Properties          map[string]string      `yaml:"properties"`
	Master              Master                 `yaml:"master"`
	Worker              Worker                 `yaml:"worker"`
	Fuse                Fuse                   `yaml:"fuse"`
	Mounts              Mounts                 `yaml:"mounts"`
	HadoopConfig        HadoopConfig           `yaml:"hadoopConfig,omitempty"`
	Secret              string                 `yaml:"secret,omitempty"`
	Tolerations         []v1.Toleration        `yaml:"tolerations,omitempty"`
	InitPortCheck       common.InitPortCheck   `yaml:"initPortCheck,omitempty"`
	LogConfig           map[string]string      `yaml:"logConfig,omitempty"`
	FuseLogConfig       map[string]string      `yaml:"fuseLogConfig,omitempty"`
	PlacementMode       string                 `yaml:"placement,omitempty"`
	Owner               *common.OwnerReference `yaml:"owner,omitempty"`
	RuntimeIdentity     common.RuntimeIdentity `yaml:"runtimeIdentity"`
	ClusterDomain       string                 `yaml:"clusterDomain,omitempty"`
	UFSVolumes          []UFSVolume            `yaml:"ufsVolumes,omitempty"`
	SecretKey           string                 `yaml:"secretKey,omitempty"`
	SecretValue         string                 `yaml:"secretValue,omitempty"`
	UseStsToken         bool                   `yaml:"UseStsToken"`
	MountType           string                 `yaml:"mountType,omitempty"`
}

type HadoopConfig struct {
	ConfigMap       string `yaml:"configMap"`
	IncludeHdfsSite bool   `yaml:"includeHdfsSite"`
	IncludeCoreSite bool   `yaml:"includeCoreSite"`
}

type Master struct {
	ReplicaCount        int                  `yaml:"replicaCount"`
	Resources           Resources            `yaml:"resources"`
	NodeSelector        map[string]string    `yaml:"nodeSelector,omitempty"`
	MasterProperties    map[string]string    `yaml:"properties"`
	FileStoreProperties map[string]string    `yaml:"fileStoreProperties"`
	TokenProperties     map[string]string    `yaml:"secretProperties"`
	Port                Ports                `yaml:"ports,omitempty"`
	OssKey              string               `yaml:"osskey,omitempty"`
	OssSecret           string               `yaml:"osssecret,omitempty"`
	Tolerations         []v1.Toleration      `yaml:"tolerations,omitempty"`
	DnsServer           string               `yaml:"dnsServer,omitempty"`
	NameSpace           string               `yaml:"namespace,omitempty"`
	Labels              map[string]string    `yaml:"labels,omitempty"`
	Annotations         map[string]string    `yaml:"annotations,omitempty"`
	ServiceCount        int                  `yaml:"svccount"`
	Env                 map[string]string    `yaml:"env,omitempty"`
	CacheSets           map[string]*CacheSet `yaml:"cachesets,omitempty"`
}

type Worker struct {
	ReplicaCount     int               `yaml:"replicaCount"`
	Resources        Resources         `yaml:"resources,omitempty"`
	NodeSelector     map[string]string `yaml:"nodeSelector,omitempty"`
	WorkerProperties map[string]string `yaml:"properties"`
	Port             Ports             `yaml:"ports,omitempty"`
	Tolerations      []v1.Toleration   `yaml:"tolerations,omitempty"`
	// Affinity         v1.Affinity       `yaml:"affinity,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Path        string            `yaml:"dataPath"`
	Env         map[string]string `yaml:"env,omitempty"`
}

type Ports struct {
	Rpc  int `yaml:"rpc,omitempty"`
	Raft int `yaml:"raft,omitempty"`
}

type Fuse struct {
	Args           []string          `yaml:"args"`
	HostPath       string            `yaml:"hostPath"`
	NodeSelector   map[string]string `yaml:"nodeSelector,omitempty"`
	FuseProperties map[string]string `yaml:"properties"`
	Global         bool              `yaml:"global,omitempty"`
	RunAs          string            `yaml:"runAs,omitempty"`
	Tolerations    []v1.Toleration   `yaml:"tolerations,omitempty"`
	Labels         map[string]string `yaml:"labels,omitempty"`
	Annotations    map[string]string `yaml:"annotations,omitempty"`
	CriticalPod    bool              `yaml:"criticalPod,omitempty"`
	Resources      Resources         `yaml:"resources,omitempty"`
	MountPath      string            `yaml:"mountPath,omitempty"`
	Mode           string            `yaml:"mode,omitempty"`
	Env            map[string]string `yaml:"env,omitempty"`
}

type Mounts struct {
	Master            map[string]*Level `yaml:"master"`
	WorkersAndClients map[string]*Level `yaml:"workersAndClients"`
}

type Resources struct {
	Limits   Resource `yaml:"limits"`
	Requests Resource `yaml:"requests"`
}

type Resource struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type Level struct {
	Path       string `yaml:"path,omitempty"`
	Type       string `yaml:"type,omitempty"`
	MediumType string `yaml:"mediumType,omitempty"`
	Quota      string `yaml:"quota,omitempty"`
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
	Name          string `yaml:"name"`
	SubPath       string `yaml:"subPath,omitempty"`
	ContainerPath string `yaml:"containerPath"`
}

type CacheSet struct {
	Name          string `yaml:"name"`
	Path          string `yaml:"path,omitempty"`
	CacheStrategy string `yaml:"cacheStrategy"`
	MetaPolicy    string `yaml:"metaPolicy"`
	ReadPolicy    string `yaml:"readPolicy"`
	WritePolicy   string `yaml:"writePolicy"`
}
