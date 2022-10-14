/*
Copyright 2022 The Fluid Author.

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

package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	v1 "k8s.io/api/core/v1"
)

type Jindo struct {
	Image           string                 `json:"image"`
	ImageTag        string                 `json:"imageTag"`
	ImagePullPolicy string                 `json:"imagePullPolicy"`
	FuseImage       string                 `json:"fuseImage"`
	FuseImageTag    string                 `json:"fuseImageTag"`
	User            int                    `json:"user"`
	Group           int                    `json:"group"`
	FsGroup         int                    `json:"fsGroup"`
	UseHostNetwork  bool                   `json:"useHostNetwork"`
	UseHostPID      bool                   `json:"useHostPID"`
	Properties      map[string]string      `json:"properties"`
	Master          Master                 `json:"master"`
	Worker          Worker                 `json:"worker"`
	Fuse            Fuse                   `json:"fuse"`
	Mounts          Mounts                 `json:"mounts"`
	HadoopConfig    HadoopConfig           `json:"hadoopConfig,omitempty"`
	Secret          string                 `json:"secret,omitempty"`
	Tolerations     []v1.Toleration        `json:"tolerations,omitempty"`
	InitPortCheck   common.InitPortCheck   `json:"initPortCheck,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	LogConfig       map[string]string      `json:"logConfig,omitempty"`
	PlacementMode   string                 `json:"placement,omitempty"`
	Owner           *common.OwnerReference `json:"owner,omitempty"`
	RuntimeIdentity common.RuntimeIdentity `json:"runtimeIdentity"`
}

type HadoopConfig struct {
	ConfigMap       string `json:"configMap"`
	IncludeHdfsSite bool   `json:"includeHdfsSite"`
	IncludeCoreSite bool   `json:"includeCoreSite"`
}

type Master struct {
	ReplicaCount     int               `json:"replicaCount"`
	Resources        Resources         `json:"resources"`
	NodeSelector     map[string]string `json:"nodeSelector,omitempty"`
	MasterProperties map[string]string `json:"properties"`
	TokenProperties  map[string]string `json:"secretProperties"`
	Port             Ports             `json:"ports,omitempty"`
	OssKey           string            `json:"osskey,omitempty"`
	OssSecret        string            `json:"osssecret,omitempty"`
	Tolerations      []v1.Toleration   `json:"tolerations,omitempty"`
	DnsServer        string            `json:"dnsServer,omitempty"`
	NameSpace        string            `json:"namespace,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
}

type Worker struct {
	Resources        Resources         `json:"resources,omitempty"`
	NodeSelector     map[string]string `json:"nodeSelector,omitempty"`
	WorkerProperties map[string]string `json:"properties"`
	Port             Ports             `json:"ports,omitempty"`
	Tolerations      []v1.Toleration   `json:"tolerations,omitempty"`
	// Affinity         v1.Affinity       `json:"affinity,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

type Ports struct {
	Rpc  int `json:"rpc,omitempty"`
	Raft int `json:"raft,omitempty"`
}

type Fuse struct {
	Args              []string          `json:"args"`
	HostPath          string            `json:"hostPath"`
	NodeSelector      map[string]string `json:"nodeSelector,omitempty"`
	FuseProperties    map[string]string `json:"properties"`
	Global            bool              `json:"global,omitempty"`
	RunAs             string            `json:"runAs,omitempty"`
	Tolerations       []v1.Toleration   `json:"tolerations,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	CriticalPod       bool              `json:"criticalPod,omitempty"`
	Resources         Resources         `json:"resources,omitempty"`
	MountPath         string            `json:"mountPath,omitempty"`
	VirtualFuseDevice bool              `json:"virtualFuseDevice"`
}

type Mounts struct {
	Master            map[string]string `json:"master"`
	WorkersAndClients map[string]string `json:"workersAndClients"`
}

type Resources struct {
	Limits   Resource `json:"limits"`
	Requests Resource `json:"requests"`
}

type Resource struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
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
