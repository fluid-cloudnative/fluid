/*

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

package alluxio

import (
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

// The value yaml file
type Alluxio struct {
	FullnameOverride string `yaml:"fullnameOverride"`

	common.ImageInfo `yaml:",inline"`
	common.UserInfo  `yaml:",inline"`

	NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	JvmOptions   []string          `yaml:"jvmOptions,omitempty"`

	Properties map[string]string `yaml:"properties,omitempty"`

	Master Master `yaml:"master,omitempty"`

	JobMaster JobMaster `yaml:"jobMaster,omitempty"`

	Worker Worker `yaml:"worker,omitempty"`

	JobWorker JobWorker `yaml:"jobWorker,omitempty"`

	Fuse Fuse `yaml:"fuse,omitempty"`

	APIGateway APIGateway `yaml:"apiGateway,omitempty"`

	TieredStore TieredStore `yaml:"tieredstore,omitempty"`

	Metastore Metastore `yaml:"metastore,omitempty"`

	Journal Journal `yaml:"journal,omitempty"`

	ShortCircuit ShortCircuit `yaml:"shortCircuit,omitempty"`
	// Enablefluid bool `yaml:"enablefluid,omitempty"`

	UFSPaths []UFSPath `yaml:"ufsPaths,omitempty"`

	UFSVolumes []UFSVolume `yaml:"ufsVolumes,omitempty"`

	InitUsers common.InitUsers `yaml:"initUsers,omitempty"`

	Monitoring string `yaml:"monitoring,omitempty"`

	HadoopConfig HadoopConfig `yaml:"hadoopConfig,omitempty"`

	Tolerations []corev1.Toleration `yaml:"tolerations,omitempty"`

	PlacementMode string `yaml:"placement,omitempty"`
}

type HadoopConfig struct {
	ConfigMap       string `yaml:"configMap"`
	IncludeHdfsSite bool   `yaml:"includeHdfsSite"`
	IncludeCoreSite bool   `yaml:"includeCoreSite"`
}

type UFSPath struct {
	HostPath  string `yaml:"hostPath"`
	UFSVolume `yaml:",inline"`
}

type UFSVolume struct {
	Name          string `yaml:"name"`
	ContainerPath string `yaml:"containerPath"`
}

type Metastore struct {
	VolumeType string `yaml:"volumeType,omitempty"`
	Size       string `yaml:"size,omitempty"`
}

type Journal struct {
	VolumeType string `yaml:"volumeType,omitempty"`
	Size       string `yaml:"size,omitempty"`
}

type ShortCircuit struct {
	Enable     bool   `yaml:"enable,omitempty"`
	Policy     string `yaml:"policy,omitempty"`
	VolumeType string `yaml:"volumeType,omitempty"`
}

type Ports struct {
	Rpc      int `yaml:"rpc,omitempty"`
	Web      int `yaml:"web,omitempty"`
	Embedded int `yaml:"embedded,omitempty"`
	Data     int `yaml:"data,omitempty"`
	Rest     int `yaml:"rest,omitempty"`
}

type APIGateway struct {
	Enabled bool  `yaml:"enabled,omitempty"`
	Ports   Ports `yaml:"ports,omitempty"`
}

type JobMaster struct {
	Ports     Ports            `yaml:"ports,omitempty"`
	Resources common.Resources `yaml:"resources,omitempty"`
}

type JobWorker struct {
	Ports     Ports            `yaml:"ports,omitempty"`
	Resources common.Resources `yaml:"resources,omitempty"`
}

type Worker struct {
	JvmOptions   []string             `yaml:"jvmOptions,omitempty"`
	Env          map[string]string    `yaml:"env,omitempty"`
	NodeSelector map[string]string    `yaml:"nodeSelector,omitempty"`
	Properties   map[string]string    `yaml:"properties,omitempty"`
	HostNetwork  bool                 `yaml:"hostNetwork,omitempty"`
	Resources    common.Resources     `yaml:"resources,omitempty"`
	Ports        Ports                `yaml:"ports,omitempty"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes      []corev1.Volume      `yaml:"volumes,omitempty"`
}

type Master struct {
	JvmOptions   []string             `yaml:"jvmOptions,omitempty"`
	Env          map[string]string    `yaml:"env,omitempty"`
	Affinity     Affinity             `yaml:"affinity"`
	NodeSelector map[string]string    `yaml:"nodeSelector,omitempty"`
	Properties   map[string]string    `yaml:"properties,omitempty"`
	Replicas     int32                `yaml:"replicaCount,omitempty"`
	HostNetwork  bool                 `yaml:"hostNetwork,omitempty"`
	Resources    common.Resources     `yaml:"resources,omitempty"`
	Ports        Ports                `yaml:"ports,omitempty"`
	BackupPath   string               `yaml:"backupPath,omitempty"`
	Restore      Restore              `yaml:"restore,omitempty"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes      []corev1.Volume      `yaml:"volumes,omitempty"`
}

type Restore struct {
	Enabled bool   `yaml:"enabled,omitempty"`
	Path    string `yaml:"path,omitempty"`
	PVCName string `yaml:"pvcName,omitempty"`
}

type Fuse struct {
	Image              string               `yaml:"image,omitempty"`
	NodeSelector       map[string]string    `yaml:"nodeSelector,omitempty"`
	ImageTag           string               `yaml:"imageTag,omitempty"`
	ImagePullPolicy    string               `yaml:"imagePullPolicy,omitempty"`
	Properties         map[string]string    `yaml:"properties,omitempty"`
	Env                map[string]string    `yaml:"env,omitempty"`
	JvmOptions         []string             `yaml:"jvmOptions,omitempty"`
	MountPath          string               `yaml:"mountPath,omitempty"`
	ShortCircuitPolicy string               `yaml:"shortCircuitPolicy,omitempty"`
	Args               []string             `yaml:"args,omitempty"`
	HostNetwork        bool                 `yaml:"hostNetwork,omitempty"`
	Enabled            bool                 `yaml:"enabled,omitempty"`
	Resources          common.Resources     `yaml:"resources,omitempty"`
	Global             bool                 `yaml:"global,omitempty"`
	CriticalPod        bool                 `yaml:"criticalPod,omitempty"`
	VolumeMounts       []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	Volumes            []corev1.Volume      `yaml:"volumes,omitempty"`
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

type Affinity struct {
	NodeAffinity *NodeAffinity `yaml:"nodeAffinity"`
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
