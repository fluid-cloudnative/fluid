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

package goosefs

import (
	"testing"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func TestGooseFSEngine_SyncRuntime(t *testing.T) {
	type fields struct {
		// runtime                *datav1alpha1.GooseFSRuntime
		// name                   string
		// namespace              string
		// runtimeType            string
		// Log                    logr.Logger
		// Client                 client.Client
		// gracefulShutdownLimits int32
		// retryShutdown          int32
		// initImage              string
		// MetadataSyncDoneCh     chan MetadataSyncResult
		// runtimeInfo            base.RuntimeInfoInterface
		// UnitTest               bool
		// lastCacheHitStates     *cacheHitStates
		// Helper                 *ctrl.Helper
		// Recorder               record.EventRecorder
	}
	type args struct {
		ctx cruntime.ReconcileRequestContext
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantChanged bool
		wantErr     bool
	}{
		{
			name:        "default",
			wantChanged: false,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				// runtime:                tt.fields.runtime,
				// name:                   tt.fields.name,
				// namespace:              tt.fields.namespace,
				// runtimeType:            tt.fields.runtimeType,
				// Log:                    tt.fields.Log,
				// Client:                 tt.fields.Client,
				// gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				// retryShutdown:          tt.fields.retryShutdown,
				// initImage:              tt.fields.initImage,
				// MetadataSyncDoneCh:     tt.fields.MetadataSyncDoneCh,
				// runtimeInfo:            tt.fields.runtimeInfo,
				// UnitTest:               tt.fields.UnitTest,
				// lastCacheHitStates:     tt.fields.lastCacheHitStates,
				// Helper:                 tt.fields.Helper,
				// Recorder:               tt.fields.Recorder,
			}
			gotChanged, err := e.SyncRuntime(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.SyncRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("GooseFSEngine.SyncRuntime() = %v, want %v", gotChanged, tt.wantChanged)
			}
		})
	}
}
