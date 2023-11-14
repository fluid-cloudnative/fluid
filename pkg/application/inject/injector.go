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

package inject

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Injector interface {
	InjectPod(in *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (out *corev1.Pod, err error)

	InjectUnstructured(in *unstructured.Unstructured, runtimeInfos map[string]base.RuntimeInfoInterface) (out *unstructured.Unstructured, err error)

	Inject(in runtime.Object, runtimeInfos map[string]base.RuntimeInfoInterface) (out runtime.Object, err error)
}
