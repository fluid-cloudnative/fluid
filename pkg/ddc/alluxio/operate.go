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

package alluxio

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *AlluxioEngine) GetDataOperationValueFile(ctx cruntime.ReconcileRequestContext, object client.Object, operation dataoperation.OperationReconcilerInterface) (valueFileName string, err error) {
	operationType := operation.GetOperationType()

	switch operationType {
	case datav1alpha1.DataBackupType:
		valueFileName, err = e.generateDataBackupValueFile(ctx, object)
		return valueFileName, err
	case datav1alpha1.DataLoadType:
		valueFileName, err = e.generateDataLoadValueFile(ctx, object)
		return valueFileName, err
	case datav1alpha1.DataProcessType:
		valueFileName, err = e.generateDataProcessValueFile(ctx, object)
		return valueFileName, err
	default:
		return "", errors.NewNotSupported(
			schema.GroupResource{
				Group:    object.GetObjectKind().GroupVersionKind().Group,
				Resource: object.GetObjectKind().GroupVersionKind().Kind,
			}, "AlluxioRuntime")
	}
}
