/*
Copyright 2021 The Fluid Authors.

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

package transfromer

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenerateOwnerReferenceFromObject(obj client.Object) *common.OwnerReference {

	ref := &common.OwnerReference{
		APIVersion:         obj.GetObjectKind().GroupVersionKind().GroupKind().Group + "/" + obj.GetObjectKind().GroupVersionKind().Version,
		Kind:               obj.GetObjectKind().GroupVersionKind().Kind,
		UID:                string(obj.GetUID()),
		Enabled:            true,
		Name:               obj.GetName(),
		BlockOwnerDeletion: false,
		Controller:         true,
	}

	return ref

}
