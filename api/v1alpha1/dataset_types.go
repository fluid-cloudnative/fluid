/*
Copyright 2022 The Fluid Authors.

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

package v1alpha1

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "github.com/rook/rook/pkg/apis/rook.io/v1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// DatasetPhase indicates whether the loading is behaving
type DatasetPhase string

const (
	// TODO: add the Pending phase to Dataset
	PendingDatasetPhase DatasetPhase = "Pending"
	// Bound to dataset, can't be released
	BoundDatasetPhase DatasetPhase = "Bound"
	// Failed, can't be deleted
	FailedDatasetPhase DatasetPhase = "Failed"
	// Not bound to runtime, can be deleted
	NotBoundDatasetPhase DatasetPhase = "NotBound"
	// updating dataset, can't be released
	UpdatingDatasetPhase DatasetPhase = "Updating"
	// migrating dataset, can't be mounted
	DataMigrating DatasetPhase = "DataMigrating"
	// the dataset have no phase and need to be judged
	NoneDatasetPhase DatasetPhase = ""
)

type SecretKeySelector struct {
	// The name of required secret
	// +required
	Name string `json:"name,omitempty"`

	// The required key in the secret
	// +optional
	Key string `json:"key,omitempty"`
}

type EncryptOptionSource struct {
	// The encryptInfo obtained from secret
	// +optional
	SecretKeyRef SecretKeySelector `json:"secretKeyRef,omitempty"`
}
type EncryptOption struct {
	// The name of encryptOption
	// +required
	Name string `json:"name,omitempty"`

	// The valueFrom of encryptOption
	// +optional
	ValueFrom EncryptOptionSource `json:"valueFrom,omitempty"`
}

// Mount describes a mounting. <br>
// Refer to <a href="https://docs.alluxio.io/os/user/stable/en/ufs/S3.html">Alluxio Storage Integrations</a> for more info
type Mount struct {
	// MountPoint is the mount point of source.
	// +kubebuilder:validation:MinLength=5
	// +required
	MountPoint string `json:"mountPoint,omitempty"`

	// The Mount Options. <br>
	// Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Mount Options</a>.  <br>
	// The option has Prefix 'fs.' And you can Learn more from
	// <a href="https://docs.alluxio.io/os/user/stable/en/ufs/S3.html">The Storage Integrations</a>
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// The name of mount
	// +kubebuilder:validation:MinLength=0
	// +optional
	Name string `json:"name,omitempty"`

	// The path of mount, if not set will be /{Name}
	// +optional
	Path string `json:"path,omitempty"`

	// Optional: Defaults to false (read-write).
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`

	// Optional: Defaults to false (shared).
	// +optional
	Shared bool `json:"shared,omitempty"`

	// The secret information
	// +optional
	EncryptOptions []EncryptOption `json:"encryptOptions,omitempty"`
}

// DataRestoreLocation describes the spec restore location of  Dataset
type DataRestoreLocation struct {
	// Path describes the path of restore, in the form of  local://subpath or pvc://<pvcName>/subpath
	// +optional
	Path string `json:"path,omitempty"`
	// NodeName describes the nodeName of restore if Path is  in the form of local://subpath
	// +optional
	NodeName string `json:"nodeName,omitempty"`
}

// DatasetSpec defines the desired state of Dataset
type DatasetSpec struct {
	// Mount Points to be mounted on Alluxio.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:UniqueItems=false
	// +required
	Mounts []Mount `json:"mounts"`

	// The owner of the dataset
	// +optional
	Owner *User `json:"owner,omitempty"`

	// NodeAffinity defines constraints that limit what nodes this dataset can be cached to.
	// This field influences the scheduling of pods that use the cached dataset.
	// +optional
	NodeAffinity *CacheableNodeAffinity `json:"nodeAffinity,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// AccessModes contains all ways the volume backing the PVC can be mounted
	// +optional
	AccessModes []v1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Runtimes for supporting dataset (e.g. AlluxioRuntime)
	Runtimes []Runtime `json:"runtimes,omitempty"`

	// Manage switch for opening Multiple datasets single node deployment or not
	// TODO(xieydd) In future, evaluate node resources and runtime resources to decide whether to turn them on
	// +kubebuilder:validation:Enum=Exclusive;"";Shared
	// +optional
	PlacementMode PlacementMode `json:"placement,omitempty"`

	// DataRestoreLocation is the location to load data of dataset  been backuped
	// +optional
	DataRestoreLocation *DataRestoreLocation `json:"dataRestoreLocation,omitempty"`

	// SharedOptions is the options to all mount
	// +optional
	SharedOptions map[string]string `json:"sharedOptions,omitempty"`

	// SharedEncryptOptions is the encryptOption to all mount
	// +optional
	SharedEncryptOptions []EncryptOption `json:"sharedEncryptOptions,omitempty"`
}

// Runtime describes a runtime to be used to support dataset
type Runtime struct {

	// Name of the runtime object
	Name string `json:"name,omitempty"`

	// Namespace of the runtime object
	Namespace string `json:"namespace,omitempty"`

	// Category the runtime object belongs to (e.g. Accelerate)
	Category common.Category `json:"category,omitempty"`

	// Runtime object's type (e.g. Alluxio)
	Type string `json:"type,omitempty"`

	// Runtime master replicas
	MasterReplicas int32 `json:"masterReplicas,omitempty"`
}

// DatasetStatus defines the observed state of Dataset
// +kubebuilder:subresource:status
type DatasetStatus struct {
	// the info of mount points have been mounted
	Mounts []Mount `json:"mounts,omitempty"`

	// Total in GB of dataset in the cluster
	UfsTotal string `json:"ufsTotal,omitempty"`

	// Dataset Phase. One of the four phases: `Pending`, `Bound`, `NotBound` and `Failed`
	Phase DatasetPhase `json:"phase,omitempty"`

	// Runtimes for supporting dataset
	Runtimes []Runtime `json:"runtimes,omitempty"`

	// Conditions is an array of current observed conditions.
	Conditions []DatasetCondition `json:"conditions"`

	// CacheStatus represents the total resources of the dataset.
	CacheStates common.CacheStateList `json:"cacheStates,omitempty"`

	// HCFSStatus represents hcfs info
	HCFSStatus *HCFSStatus `json:"hcfs,omitempty"`

	// FileNum represents the file numbers of the dataset
	FileNum string `json:"fileNum,omitempty"`

	// DataLoadRef specifies the running DataLoad job that targets this Dataset.
	// This is mainly used as a lock to prevent concurrent DataLoad jobs.
	// Deprecated, use OperationRef instead
	DataLoadRef string `json:"dataLoadRef,omitempty"`

	// DataBackupRef specifies the running Backup job that targets this Dataset.
	// This is mainly used as a lock to prevent concurrent DataBackup jobs.
	// Deprecated, use OperationRef instead
	DataBackupRef string `json:"dataBackupRef,omitempty"`

	// OperationRef specifies the Operation that targets this Dataset.
	// This is mainly used as a lock to prevent concurrent same Operation jobs.
	OperationRef map[string]string `json:"operationRef,omitempty"`

	// DatasetRef specifies the datasets namespaced name mounting this Dataset.
	DatasetRef []string `json:"datasetRef,omitempty"`
}

// DatasetConditionType defines all kinds of types of cacheStatus.<br>
// one of the three types: `RuntimeScheduled`, `Ready` and `Initialized`
type DatasetConditionType string

const (
	// RuntimeScheduled means the runtime CRD has been accepted by the system,
	// But master and workers are not ready
	RuntimeScheduled DatasetConditionType = "RuntimeScheduled"

	// DatasetReady means the cache system for the dataset is ready.
	DatasetReady DatasetConditionType = "Ready"

	// DatasetNotReady means the dataset is not bound due to some unexpected error
	DatasetNotReady DatasetConditionType = "NotReady"

	// DatasetUpdateReady means the cache system for the dataset is updated.
	DatasetUpdateReady DatasetConditionType = "UpdateReady"

	// DatasetUpdating means the cache system for the dataset is updating.
	DatasetUpdating DatasetConditionType = "Updating"

	// DatasetInitialized means the cache system for the dataset is Initialized.
	DatasetInitialized DatasetConditionType = "Initialized"
)

// CacheableNodeAffinity defines constraints that limit what nodes this dataset can be cached to.
type CacheableNodeAffinity struct {
	// Required specifies hard node constraints that must be met.
	Required *v1.NodeSelector `json:"required,omitempty"`
}

// Condition describes the state of the cache at a certain point.
type DatasetCondition struct {
	// Type of cache condition.
	Type DatasetConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +kubebuilder:printcolumn:name="Ufs Total Size",type="string",JSONPath=`.status.ufsTotal`
// +kubebuilder:printcolumn:name="Cached",type="string",JSONPath=`.status.cacheStates.cached`
// +kubebuilder:printcolumn:name="Cache Capacity",type="string",JSONPath=`.status.cacheStates.cacheCapacity`
// +kubebuilder:printcolumn:name="Cached Percentage",type="string",JSONPath=`.status.cacheStates.cachedPercentage`
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="HCFS URL",type="string",JSONPath=`.status.hcfs.endpoint`,priority=10
// +kubebuilder:printcolumn:name="TOTAL FILES",type="string",JSONPath=`.status.fileNum`,priority=11
// +kubebuilder:printcolumn:name="CACHE HIT RATIO",type="string",JSONPath=`.status.cacheStates.cacheHitRatio`,priority=10
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid},shortName=dataset
// +genclient

// Dataset is the Schema for the datasets API
type Dataset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatasetSpec   `json:"spec,omitempty"`
	Status DatasetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// DatasetList contains a list of Dataset
type DatasetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dataset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dataset{}, &DatasetList{})
}

// CanbeBound checks if the dataset can be bound to the runtime
func (dataset *Dataset) CanbeBound(name string, namespace string, category common.Category) (bound bool) {

	if len(dataset.Status.Runtimes) == 0 {
		bound = true
	}

	for _, runtime := range dataset.Status.Runtimes {
		if runtime.Name == name &&
			runtime.Namespace == namespace &&
			runtime.Category == category {
			bound = true
		}
	}

	return bound
}

func (dataset *Dataset) IsExclusiveMode() bool {
	return dataset.Spec.PlacementMode == DefaultMode || dataset.Spec.PlacementMode == ExclusiveMode
}

// GetDataOperationInProgress get the name of operation for certain type running on this dataset, otherwise return empty string
func (dataset *Dataset) GetDataOperationInProgress(operationType string) string {
	if dataset.Status.OperationRef == nil {
		return ""
	}

	return dataset.Status.OperationRef[operationType]
}

// SetDataOperationInProgress set the data operation running on this dataset,
func (dataset *Dataset) SetDataOperationInProgress(operationType string, name string) {
	if dataset.Status.OperationRef == nil {
		dataset.Status.OperationRef = map[string]string{
			operationType: name,
		}
		return
	}
	if dataset.Status.OperationRef[operationType] == "" {
		dataset.Status.OperationRef[operationType] = name
		return
	}

	dataset.Status.OperationRef[operationType] = dataset.Status.OperationRef[operationType] + "," + name
}

// RemoveDataOperationInProgress release Dataset for operation
func (dataset *Dataset) RemoveDataOperationInProgress(operationType, name string) string {
	if dataset.Status.OperationRef == nil {
		return ""
	}
	dataOpKeys := strings.Split(dataset.Status.OperationRef[operationType], ",")
	if len(dataOpKeys) == 1 && dataOpKeys[0] == name {
		delete(dataset.Status.OperationRef, operationType)
		return ""
	}
	for i, key := range dataOpKeys {
		if key == name {
			dataOpKeys = append(dataOpKeys[:i], dataOpKeys[i+1:]...)
			break
		}
	}
	dataset.Status.OperationRef[operationType] = strings.Join(dataOpKeys, ",")
	return dataset.Status.OperationRef[operationType]
}
