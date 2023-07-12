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

// DataProcessSpec defines the desired state of DataProcess
type DataProcessSpec struct {
	
	// Foo is an example field of DataProcess. Edit dataprocess_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DataProcess is the Schema for the dataprocesses API
type DataProcess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataProcessSpec `json:"spec,omitempty"`
	Status OperationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataProcessList contains a list of DataProcess
type DataProcessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataProcess `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataProcess{}, &DataProcessList{})
}
