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
	"time"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Engine = (*TemplateEngine)(nil)

// TemplateEngine implements the Engine interface and some default methods.
type TemplateEngine struct {
	Implement
	// Embed the default implementation to satisfy ExtendedLifecycleManager interface
	DefaultExtendedLifecycleManager

	Id          string
	client.Client
	Log         logr.Logger
	Context     cruntime.ReconcileRequestContext
	runtimeType string

	// Fields required by syncs.go logic
	timeOfLastSync time.Time
	permitSync     bool
}

func NewTemplateEngine(impl Implement, id string, ctx cruntime.ReconcileRequestContext) *TemplateEngine {
	return &TemplateEngine{
		Implement:                       impl,
		Id:                              id,
		Client:                          ctx.Client,
		Log:                             ctx.Log,
		Context:                         ctx,
		runtimeType:                     ctx.RuntimeType,
		DefaultExtendedLifecycleManager: DefaultExtendedLifecycleManager{},
		
		// Initialize sync fields
		permitSync:     true,
		timeOfLastSync: time.Now().Add(-1 * time.Hour), // Ensure immediate first sync
	}
}

// ID returns the id of the engine
func (t *TemplateEngine) ID() string {
	return t.Id
}

// NOTE: Sync() and Operate() methods are deliberately omitted here
// because they are defined in pkg/ddc/base/syncs.go and pkg/ddc/base/operation.go