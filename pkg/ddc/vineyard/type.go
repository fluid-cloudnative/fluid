/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// The value yaml file
type Vineyard struct {
	FullnameOverride string `json:"fullnameOverride"`

	common.ImageInfo `json:",inline"`
	common.UserInfo  `json:",inline"`

	Master `json:"master,omitempty"`
	Worker `json:"worker,omitempty"`

	Owner *common.OwnerReference `json:"owner,omitempty"`

	Fuse `json:"fuse,omitempty"`

	TieredStore TieredStore `json:"tieredstore,omitempty"`

	DisablePrometheus bool            `json:"disablePrometheus,omitempty"`
	Volumes           []corev1.Volume `json:"volumes,omitempty"`
}

type Endpoint struct {
	URI            string                       `json:"uri,omitempty"`
	EncryptOptions []datav1alpha1.EncryptOption `json:"encryptOptions,omitempty"`
	Options        map[string]string            `json:"options,omitempty"`
}

type Master struct {
	Replicas         int32                `json:"replicas,omitempty"`
	Image            string               `json:"image,omitempty"`
	ImageTag         string               `json:"imageTag,omitempty"`
	ImagePullPolicy  string               `json:"imagePullPolicy,omitempty"`
	NodeSelector     map[string]string    `json:"nodeSelector,omitempty"`
	Ports            map[string]int       `json:"ports,omitempty"`
	Env              map[string]string    `json:"env,omitempty"`
	Options          map[string]string    `json:"options,omitempty"`
	Resources        common.Resources     `json:"resources,omitempty"`
	VolumeMounts     []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	ExternalEndpoint Endpoint             `json:"externalEndpoint,omitempty"`
}

type Worker struct {
	Replicas        int32                `json:"replicas,omitempty"`
	Image           string               `json:"image,omitempty"`
	ImageTag        string               `json:"imageTag,omitempty"`
	ImagePullPolicy string               `json:"imagePullPolicy,omitempty"`
	NodeSelector    map[string]string    `json:"nodeSelector,omitempty"`
	Ports           map[string]int       `json:"ports,omitempty"`
	Env             map[string]string    `json:"env,omitempty"`
	Options         map[string]string    `json:"options,omitempty"`
	Resources       common.Resources     `json:"resources,omitempty"`
	VolumeMounts    []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

type Fuse struct {
	Image           string                       `json:"image,omitempty"`
	ImageTag        string                       `json:"imageTag,omitempty"`
	ImagePullPolicy string                       `json:"imagePullPolicy,omitempty"`
	Env             map[string]string            `json:"env,omitempty"`
	CleanPolicy     datav1alpha1.FuseCleanPolicy `json:"cleanPolicy,omitempty"`
	TargetPath      string                       `json:"targetPath,omitempty"`
	NodeSelector    map[string]string            `json:"nodeSelector,omitempty"`
	Resources       common.Resources             `json:"resources,omitempty"`
	HostPID         bool                         `json:"hostPID,omitempty"`
}

type TieredStore struct {
	Levels []Level `json:"levels,omitempty"`
}

type Level struct {
	MediumType common.MediumType `json:"mediumtype,omitempty"`

	VolumeType common.VolumeType `json:"volumetype,omitempty"`

	VolumeSource datav1alpha1.VolumeSource `json:"volumesource,omitempty"`

	Level int `json:"level,omitempty"`

	Path string `json:"path,omitempty"`

	Quota *resource.Quantity `json:"quota,omitempty"`

	QuotaList string `json:"quotaList,omitempty"`

	High string `json:"high,omitempty"`

	Low string `json:"low,omitempty"`
}

type cacheStates struct {
	cacheCapacity    string
	cached           string
	cachedPercentage string
}
