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

import "github.com/fluid-cloudnative/fluid/pkg/common"

// The value yaml file
type Alluxio struct {
	FullnameOverride string `yaml:"fullnameOverride"`

	ImageInfo `yaml:",inline"`
	UserInfo  `yaml:",inline"`

	NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	JvmOptions   []string          `yaml:"jvmOptions,omitempty"`

	Properties map[string]string `yaml:"properties,omitempty"`

	Master Master `yaml:"master,omitempty"`

	Worker Worker `yaml:"worker,omitempty"`

	Fuse Fuse `yaml:"fuse,omitempty"`

	Tieredstore Tieredstore `yaml:"tieredstore,omitempty"`

	Metastore Metastore `yaml:"metastore,omitempty"`

	Journal Journal `yaml:"journal,omitempty"`

	ShortCircuit ShortCircuit `yaml:"shortCircuit,omitempty"`
	// Enablefluid bool `yaml:"enablefluid,omitempty"`
}

type ImageInfo struct {
	Image           string `yaml:"image"`
	ImageTag        string `yaml:"imageTag"`
	ImagePullPolicy string `yaml:"imagePullPolicy"`
}

type UserInfo struct {
	User    int `yaml:"user"`
	Group   int `yaml:"group"`
	FSGroup int `yaml:"fsGroup"`
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

type Worker struct {
	JvmOptions   []string          `yaml:"jvmOptions,omitempty"`
	Env          map[string]string `yaml:"env,omitempty"`
	NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	Properties   map[string]string `yaml:"properties,omitempty"`
	HostNetwork  bool              `yaml:"hostNetwork,omitempty"`
	Resources    common.Resources  `yaml:"resources,omitempty"`
}

type Master struct {
	JvmOptions   []string          `yaml:"jvmOptions,omitempty"`
	Env          map[string]string `yaml:"env,omitempty"`
	NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	Properties   map[string]string `yaml:"properties,omitempty"`
	Replicas     int32             `yaml:"replicaCount,omitempty"`
	HostNetwork  bool              `yaml:"hostNetwork,omitempty"`
	Resources    common.Resources  `yaml:"resources,omitempty"`
}

type Fuse struct {
	Image              string            `yaml:"image,omitempty"`
	NodeSelector       map[string]string `yaml:"nodeSelector,omitempty"`
	ImageTag           string            `yaml:"imageTag,omitempty"`
	ImagePullPolicy    string            `yaml:"imagePullPolicy,omitempty"`
	Properties         map[string]string `yaml:"properties,omitempty"`
	Env                map[string]string `yaml:"env,omitempty"`
	JvmOptions         []string          `yaml:"jvmOptions,omitempty"`
	MountPath          string            `yaml:"mountPath,omitempty"`
	ShortCircuitPolicy string            `yaml:"shortCircuitPolicy,omitempty"`
	Args               []string          `yaml:"args,omitempty"`
	HostNetwork        bool              `yaml:"hostNetwork,omitempty"`
	Enabled            bool              `yaml:"enabled,omitempty"`
	Resources          common.Resources  `yaml:"resources,omitempty"`
}

type Tieredstore struct {
	Levels []Level `yaml:"levels,omitempty"`
}

type Level struct {
	Alias      string `yaml:"alias,omitempty"`
	Level      int    `yaml:"level"`
	Mediumtype string `yaml:"mediumtype,omitempty"`
	Type       string `yaml:"type,omitempty"`
	Path       string `yaml:"path,omitempty"`
	Quota      string `yaml:"quota,omitempty"`
	High       string `yaml:"high,omitempty"`
	Low        string `yaml:"low,omitempty"`
}

type cacheStates struct {
	cacheCapacity string
	// cacheable        string
	// lowWaterMark     string
	// highWaterMark    string
	cached           string
	cachedPercentage string
	nonCacheable     string
}

func (value *Alluxio) getTiredStoreLevel0Path() (path string) {
	path = "/dev/shm"
	for _, level := range value.Tieredstore.Levels {
		if level.Level == 0 {
			path = level.Path
			break
		}
	}
	return
}
