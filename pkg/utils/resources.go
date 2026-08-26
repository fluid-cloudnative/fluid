/*
Copyright 2023 The Fluid Authors.

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

package utils

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/pkg/errors"
)

func TransformCoreV1ResourcesToInternalResources(res corev1.ResourceRequirements) (cRes common.Resources) {

	cRes = common.Resources{
		Requests: make(common.ResourceList, len(res.Requests)),
		Limits:   make(common.ResourceList, len(res.Limits)),
	}

	if len(res.Requests) > 0 {
		for k, v := range res.Requests {
			cRes.Requests[k] = v.String()
		}
	}

	if len(res.Limits) > 0 {
		for k, v := range res.Limits {
			cRes.Limits[k] = v.String()
		}
	}

	return
}

// TransformInternalResourcesToCoreV1Resources is the inverse of
// TransformCoreV1ResourcesToInternalResources. It is needed when a value rendered into a Helm
// values ConfigMap has to be compared against or written back to a live workload spec.
func TransformInternalResourcesToCoreV1Resources(cRes common.Resources) (res corev1.ResourceRequirements, err error) {
	res.Requests, err = transformInternalResourceListToCoreV1ResourceList(cRes.Requests)
	if err != nil {
		return res, errors.Wrap(err, "failed to parse resource requests")
	}

	res.Limits, err = transformInternalResourceListToCoreV1ResourceList(cRes.Limits)
	if err != nil {
		return res, errors.Wrap(err, "failed to parse resource limits")
	}

	return res, nil
}

func transformInternalResourceListToCoreV1ResourceList(cList common.ResourceList) (list corev1.ResourceList, err error) {
	if len(cList) == 0 {
		return nil, nil
	}

	list = make(corev1.ResourceList, len(cList))
	for k, v := range cList {
		quantity, parseErr := resource.ParseQuantity(v)
		if parseErr != nil {
			return nil, errors.Wrapf(parseErr, "failed to parse quantity %q of resource %q", v, k)
		}
		list[k] = quantity
	}

	return list, nil
}

func ResourceRequirementsEqual(source corev1.ResourceRequirements,
	target corev1.ResourceRequirements) bool {
	return resourceListsEqual(source.Requests, target.Requests) &&
		resourceListsEqual(source.Limits, target.Limits)
}

func resourceListsEqual(a corev1.ResourceList, b corev1.ResourceList) bool {
	a = withoutZeroElems(a)
	b = withoutZeroElems(b)
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		vb, found := b[k]
		if !found {
			return false
		}
		if v.Cmp(vb) != 0 {
			return false
		}
	}
	return true
}

func withoutZeroElems(input corev1.ResourceList) (output corev1.ResourceList) {
	output = corev1.ResourceList{}
	for k, v := range input {
		if !v.IsZero() {
			output[k] = v
		}
	}
	return
}
