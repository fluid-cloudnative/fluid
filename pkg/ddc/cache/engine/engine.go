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

package engine

import (
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/componenthelper"
	fluiderrors "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

const (
	syncRetryDurationEnv     string = "FLUID_SYNC_RETRY_DURATION"
	defaultSyncRetryDuration        = 5 * time.Second
)

// Use compiler to check if the struct implements all the interface
var _ base.Engine = (*CacheEngine)(nil)

// CacheEngine is used for handling datasets mounting another dataset.
// We use `virtual` dataset/runtime to represent the reference dataset/runtime itself,
// and use `physical` dataset/runtime to represent the dataset/runtime is mounted by virtual dataset.
type CacheEngine struct {
	client.Client
	Scheme *runtime.Scheme

	Log      logr.Logger
	Recorder record.EventRecorder

	Id        string
	name      string
	namespace string

	syncRetryDuration      time.Duration
	timeOfLastSync         time.Time
	retryShutdown          int32
	gracefulShutdownLimits int32

	masterHelper componenthelper.ComponentHelper
	workerHelper componenthelper.ComponentHelper
	clientHelper componenthelper.ComponentHelper
	value        *common.CacheRuntimeValue

	runtimeType string
	engineImpl  string
}

func (e *CacheEngine) Operate(ctx cruntime.ReconcileRequestContext, opStatus *v1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	object := operation.GetOperationObject()
	// cache runtime engine not support data operation
	err := fluiderrors.NewNotSupported(
		schema.GroupResource{
			Group:    object.GetObjectKind().GroupVersionKind().Group,
			Resource: object.GetObjectKind().GroupVersionKind().Kind,
		}, "CacheRuntime")
	ctx.Log.Error(err, "CacheRuntime does not support data operations")
	ctx.Recorder.Event(object, v1.EventTypeWarning, common.DataOperationNotSupport, "cacheEngine does not support data operations")
	return utils.NoRequeue()
}

// ID returns the id of the engine
func (e *CacheEngine) ID() string {
	return e.Id
}

func Build(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	engine := &CacheEngine{
		Id:                     id,
		name:                   ctx.Name,
		namespace:              ctx.Namespace,
		runtimeType:            ctx.RuntimeType,
		engineImpl:             ctx.EngineImpl,
		Client:                 ctx.Client,
		Scheme:                 ctx.Scheme,
		Recorder:               ctx.Recorder,
		Log:                    ctx.Log,
		gracefulShutdownLimits: 5,
		retryShutdown:          0,
	}
	engine.Log = ctx.Log.WithValues("cache engine", ctx.RuntimeType).WithValues("id", id)

	// set sync duration
	duration, err := getSyncRetryDuration()
	if err != nil {
		engine.Log.Error(err, "Failed to parse syncRetryDurationEnv: FLUID_SYNC_RETRY_DURATION, use the default setting")
	}
	if duration != nil {
		engine.syncRetryDuration = *duration
	} else {
		engine.syncRetryDuration = defaultSyncRetryDuration
	}

	engine.timeOfLastSync = time.Now().Add(-engine.syncRetryDuration)
	engine.Log.Info("Set the syncRetryDuration", "syncRetryDuration", engine.syncRetryDuration)

	return engine, err
}
