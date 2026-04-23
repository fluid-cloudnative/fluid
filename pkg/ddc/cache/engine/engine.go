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
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
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

	runtimeType string
	engineImpl  string

	// TODO(cache runtime): use narrowed interface, and as a part of RuntimeInfoInterface.
	// always use getRuntimeInfo() method instead of use this directly.
	runtimeInfo base.RuntimeInfoInterface
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

// Precheck checks if the given key can be found in the current runtime types
func Precheck(client client.Client, key types.NamespacedName) (found bool, err error) {
	var obj datav1alpha1.CacheRuntime
	return utils.CheckObject(client, key, &obj)
}
