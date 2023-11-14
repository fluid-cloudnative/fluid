/*
Copyright The Kubernetes Authors.

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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// FlexPersistentVolumeSourceApplyConfiguration represents an declarative configuration of the FlexPersistentVolumeSource type for use
// with apply.
type FlexPersistentVolumeSourceApplyConfiguration struct {
	Driver    *string                            `json:"driver,omitempty"`
	FSType    *string                            `json:"fsType,omitempty"`
	SecretRef *SecretReferenceApplyConfiguration `json:"secretRef,omitempty"`
	ReadOnly  *bool                              `json:"readOnly,omitempty"`
	Options   map[string]string                  `json:"options,omitempty"`
}

// FlexPersistentVolumeSourceApplyConfiguration constructs an declarative configuration of the FlexPersistentVolumeSource type for use with
// apply.
func FlexPersistentVolumeSource() *FlexPersistentVolumeSourceApplyConfiguration {
	return &FlexPersistentVolumeSourceApplyConfiguration{}
}

// WithDriver sets the Driver field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Driver field is set to the value of the last call.
func (b *FlexPersistentVolumeSourceApplyConfiguration) WithDriver(value string) *FlexPersistentVolumeSourceApplyConfiguration {
	b.Driver = &value
	return b
}

// WithFSType sets the FSType field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FSType field is set to the value of the last call.
func (b *FlexPersistentVolumeSourceApplyConfiguration) WithFSType(value string) *FlexPersistentVolumeSourceApplyConfiguration {
	b.FSType = &value
	return b
}

// WithSecretRef sets the SecretRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SecretRef field is set to the value of the last call.
func (b *FlexPersistentVolumeSourceApplyConfiguration) WithSecretRef(value *SecretReferenceApplyConfiguration) *FlexPersistentVolumeSourceApplyConfiguration {
	b.SecretRef = value
	return b
}

// WithReadOnly sets the ReadOnly field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ReadOnly field is set to the value of the last call.
func (b *FlexPersistentVolumeSourceApplyConfiguration) WithReadOnly(value bool) *FlexPersistentVolumeSourceApplyConfiguration {
	b.ReadOnly = &value
	return b
}

// WithOptions puts the entries into the Options field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Options field,
// overwriting an existing map entries in Options field with the same key.
func (b *FlexPersistentVolumeSourceApplyConfiguration) WithOptions(entries map[string]string) *FlexPersistentVolumeSourceApplyConfiguration {
	if b.Options == nil && len(entries) > 0 {
		b.Options = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Options[k] = v
	}
	return b
}
