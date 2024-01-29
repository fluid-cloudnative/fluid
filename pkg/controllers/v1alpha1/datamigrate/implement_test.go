/*
  Copyright 2024 The Fluid Authors.

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

package datamigrate

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestDataMigrateOperation_Validate(t *testing.T) {
	type fields struct {
		Client      client.Client
		Log         logr.Logger
		Recorder    record.EventRecorder
		dataMigrate *datav1alpha1.DataMigrate
	}
	type args struct {
		ctx runtime.ReconcileRequestContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		reason  string
		wantErr bool
	}{
		{
			name: "ssh secret not set when parallel migrate",
			fields: fields{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Parallelism:     2,
						ParallelOptions: map[string]string{},
					},
				},
			},
			args: args{
				ctx: runtime.ReconcileRequestContext{
					Dataset: nil,
				},
			},
			reason:  common.TargetSSHSecretNameNotSet,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &dataMigrateOperation{
				Client:      tt.fields.Client,
				Log:         tt.fields.Log,
				Recorder:    tt.fields.Recorder,
				dataMigrate: tt.fields.dataMigrate,
			}
			got, err := r.Validate(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && got[0].Reason != tt.reason {
				t.Errorf("Validate() error reason got = %v, want %v", got[0].Reason, tt.reason)
			}
		})
	}
}
