/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
