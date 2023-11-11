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

package thin

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func TestThinEngine_SyncRuntime(t1 *testing.T) {
	type fields struct{}
	type args struct {
		ctx runtime.ReconcileRequestContext
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
		t1.Run(tt.name, func(t1 *testing.T) {
			t := ThinEngine{}
			gotChanged, err := t.SyncRuntime(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t1.Errorf("SyncRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t1.Errorf("SyncRuntime() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
		})
	}
}
