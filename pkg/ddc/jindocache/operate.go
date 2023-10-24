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

package jindocache

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (e *JindoCacheEngine) GetDataOperationValueFile(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationInterface) (valueFileName string, err error) {
	operationType := operation.GetOperationType()
	object := operation.GetOperationObject()

	if operationType == datav1alpha1.DataLoadType {
		valueFileName, err = e.generateDataLoadValueFile(ctx, object)
		return valueFileName, err
	}

	return "", errors.NewNotSupported(
		schema.GroupResource{
			Group:    object.GetObjectKind().GroupVersionKind().Group,
			Resource: object.GetObjectKind().GroupVersionKind().Kind,
		}, "JindoRuntime")
}
