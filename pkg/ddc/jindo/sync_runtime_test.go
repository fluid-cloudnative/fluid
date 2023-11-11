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

package jindo

import (
	"testing"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func TestJindoEngine_SyncRuntime(t *testing.T) {
	type fields struct {
		// runtime                *datav1alpha1.JindoRuntime
		// name                   string
		// namespace              string
		// runtimeType            string
		// Log                    logr.Logger
		// Client                 client.Client
		// gracefulShutdownLimits int32
		// retryShutdown          int32
		// runtimeInfo            base.RuntimeInfoInterface
		// MetadataSyncDoneCh     chan MetadataSyncResult
		// cacheNodeNames         []string
		// Recorder               record.EventRecorder
		// Helper                 *ctrl.Helper
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
		// TODO: Add test cases.
		{
			name:        "default",
			wantChanged: false,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JindoEngine{
				// runtime:                tt.fields.runtime,
				// name:                   tt.fields.name,
				// namespace:              tt.fields.namespace,
				// runtimeType:            tt.fields.runtimeType,
				// Log:                    tt.fields.Log,
				// Client:                 tt.fields.Client,
				// gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				// retryShutdown:          tt.fields.retryShutdown,
				// runtimeInfo:            tt.fields.runtimeInfo,
				// MetadataSyncDoneCh:     tt.fields.MetadataSyncDoneCh,
				// cacheNodeNames:         tt.fields.cacheNodeNames,
				// Recorder:               tt.fields.Recorder,
				// Helper:                 tt.fields.Helper,
			}
			gotChanged, err := e.SyncRuntime(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("JindoEngine.SyncRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("JindoEngine.SyncRuntime() = %v, want %v", gotChanged, tt.wantChanged)
			}
		})
	}
}
