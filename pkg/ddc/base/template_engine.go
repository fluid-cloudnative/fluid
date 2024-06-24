/*
Copyright 2020 The Fluid Authors.

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

package base

import (
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/types"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	syncRetryDurationEnv string = "FLUID_SYNC_RETRY_DURATION"

	defaultSyncRetryDuration time.Duration = time.Duration(5 * time.Second)
)

// Use compiler to check if the struct implements all the interface
var _ Engine = (*TemplateEngine)(nil)

type TemplateEngine struct {
	Implement
	Id string
	client.Client
	Log               logr.Logger
	Context           cruntime.ReconcileRequestContext
	syncRetryDuration time.Duration
	timeOfLastSync    time.Time
}

// NewTemplateEngine creates template engine
func NewTemplateEngine(impl Implement,
	id string,
	// client client.Client,
	// log logr.Logger,
	context cruntime.ReconcileRequestContext) *TemplateEngine {
	b := &TemplateEngine{
		Implement: impl,
		Id:        id,
		Context:   context,
		Client:    context.Client,
		// Log:       log,
	}
	b.Log = context.Log.WithValues("engine", context.RuntimeType).WithValues("id", id)
	// b.timeOfLastSync = time.Now()
	duration, err := getSyncRetryDuration()
	if err != nil {
		b.Log.Error(err, "Failed to parse syncRetryDurationEnv: FLUID_SYNC_RETRY_DURATION, use the default setting")
	}
	if duration != nil {
		b.syncRetryDuration = *duration
	} else {
		b.syncRetryDuration = defaultSyncRetryDuration
	}
	b.timeOfLastSync = time.Now().Add(-b.syncRetryDuration)
	b.Log.Info("Set the syncRetryDuration", "syncRetryDuration", b.syncRetryDuration)

	return b
}

// ID returns the id of the engine
func (t *TemplateEngine) ID() string {
	return t.Id
}

// Shutdown and clean up the engine
func (t *TemplateEngine) Shutdown() error {
	return t.Implement.Shutdown()
}

func getSyncRetryDuration() (d *time.Duration, err error) {
	if value, existed := os.LookupEnv(syncRetryDurationEnv); existed {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return d, err
		}
		d = &duration
	}
	return
}

func (t *TemplateEngine) permitSync(key types.NamespacedName) (permit bool) {
	if time.Since(t.timeOfLastSync) < t.syncRetryDuration {
		info := fmt.Sprintf("Skipping engine.Sync(). Not permmitted until  %v (syncRetryDuration %v) since timeOfLastSync %v.",
			t.timeOfLastSync.Add(t.syncRetryDuration),
			t.syncRetryDuration,
			t.timeOfLastSync)
		t.Log.Info(info, "name", key.Name, "namespace", key.Namespace)
	} else {
		permit = true
		info := fmt.Sprintf("Processing engine.Sync(). permmitted  %v (syncRetryDuration %v) since timeOfLastSync %v.",
			t.timeOfLastSync.Add(t.syncRetryDuration),
			t.syncRetryDuration,
			t.timeOfLastSync)
		t.Log.V(1).Info(info, "name", key.Name, "namespace", key.Namespace)
	}

	return
}
