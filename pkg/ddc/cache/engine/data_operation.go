/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (e *CacheEngine) Operate(ctx cruntime.ReconcileRequestContext, opStatus *v1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	// TODO(cache runtime): Implement
	object := operation.GetOperationObject()
	// cache runtime engine current not support any data operation
	err := errors.NewNotSupported(
		schema.GroupResource{
			Group:    object.GetObjectKind().GroupVersionKind().Group,
			Resource: object.GetObjectKind().GroupVersionKind().Kind,
		}, "CacheRuntime")
	ctx.Log.Error(err, "CacheRuntime does not support data operations")
	ctx.Recorder.Event(object, v1.EventTypeWarning, common.DataOperationNotSupport, "cacheEngine does not support data operations")
	return utils.NoRequeue()
}

func (e *CacheEngine) GetDataOperationValueFile(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationInterface) (valueFileName string, err error) {
	// TODO(cache runtime): Implement
	return "", nil
}
