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

package fake

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// NewFakeClientWithScheme is to fix the issue by wrappering it:
// fake.NewFakeClientWithScheme is deprecated: Please use NewClientBuilder instead.  (staticcheck)
func NewFakeClientWithScheme(clientScheme *runtime.Scheme, initObjs ...runtime.Object) client.Client {
	var clientObjs []client.Object
	for _, obj := range initObjs {
		clientObj, ok := obj.(client.Object)
		if ok {
			clientObjs = append(clientObjs, clientObj)
		}
	}

	return fake.NewClientBuilder().WithScheme(clientScheme).WithRuntimeObjects(initObjs...).WithStatusSubresource(clientObjs...).Build()
}

// NewFakeClient is to fix the issue by wrappering it:
// fake.NewFakeClient is deprecated: Please use NewClientBuilder instead.  (staticcheck)
func NewFakeClient(initObjs ...runtime.Object) client.Client {
	return fake.NewClientBuilder().WithRuntimeObjects(initObjs...).Build()
}
