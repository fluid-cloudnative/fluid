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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataTableSpec defines the desired state of DataTable
type DataTableSpec struct {
	// Hive URL
	Url string `json:"url,omitempty"`

	// Indicates the table data to be cached
	Schemas []Schema `json:"schemas,omitempty"`
}

// Schema defines the table data to be cached in a schema
type Schema struct {
	// DataBase Name
	SchemaName string `json:"schemaName"`

	// Indicates the table to be cached
	Tables []Table `json:"tables,omitempty"`
}

// Schema defines the table data to be cached in a table
type Table struct {
	// Table Name
	TableName string `json:"tableName"`

	// Column name for this table
	ColumnName []string `json:"columnName,omitempty"`

	// Partition infos for the partition
	PartitionColumn []map[string]string `json:"partitionColumn,omitempty"` // 每个map表示一个分区表（可能会有多个kv）
}

// DataTableStatus defines the observed state of DataTable
// +kubebuilder:subresource:status
type DataTableStatus struct {
	// the data of mount points have been mounted
	Schemas []Schema `json:"mounts,omitempty"`

	// Total in GB of data in the cluster
	UfsTotal string `json:"ufsTotal,omitempty"`

	// DataTable Phase. One of the three phases: `Bound`, `NotBound` and `Failed`
	Phase DataTablePhase `json:"phase,omitempty"` // 表明数据是否挂载
}

// DataTablePhase indicates whether the loading is behaving
type DataTablePhase string

const (
	// Bound to dataset, can't be released
	BoundDataTablePhase DataTablePhase = "Bound"
	// Failed, can't be deleted
	FailedDataTablePhase DataTablePhase = "Failed"
	// Not bound to runtime, can be deleted
	NotBoundDataTablePhase DataTablePhase = "NotBound"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DataTable is the Schema for the datatables API
type DataTable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataTableSpec   `json:"spec,omitempty"`
	Status DataTableStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataTableList contains a list of DataTable
type DataTableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataTable `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataTable{}, &DataTableList{})
}
