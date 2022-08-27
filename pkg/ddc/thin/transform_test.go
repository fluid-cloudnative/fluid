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

package thin

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestThinEngine_transformTolerations(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	type args struct {
		dataset *datav1alpha1.Dataset
		value   *ThinValue
	}
	var tests = []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test",
			fields: fields{
				name:      "",
				namespace: "",
			},
			args: args{
				dataset: &datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "a",
						Operator: corev1.TolerationOpEqual,
						Value:    "b",
					}},
				}},
				value: &ThinValue{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &ThinEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			j.transformTolerations(tt.args.dataset, tt.args.value)
			if len(tt.args.value.Tolerations) != len(tt.args.dataset.Spec.Tolerations) {
				t.Errorf("transformTolerations() tolerations = %v", tt.args.value.Tolerations)
			}
		})
	}
}

func TestThinEngine_parseFromProfile(t1 *testing.T) {
	type args struct {
		profile *datav1alpha1.ThinRuntimeProfile
		value   *ThinValue
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "image",
			args: args{
				profile: nil,
				value:   nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThinEngine{
				Log: fake.NullLogger(),
			}
			if err := t.parseFromProfile(tt.args.profile, tt.args.value); (err != nil) != tt.wantErr {
				t1.Errorf("parseFromProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThinEngine_parseWorkerImage(t1 *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.ThinRuntime
		name                   string
		namespace              string
		runtimeType            string
		Log                    logr.Logger
		Client                 client.Client
		gracefulShutdownLimits int32
		MetadataSyncDoneCh     chan MetadataSyncResult
		runtimeInfo            base.RuntimeInfoInterface
		UnitTest               bool
		retryShutdown          int32
		Helper                 *ctrl.Helper
	}
	type args struct {
		runtime *datav1alpha1.ThinRuntime
		value   *ThinValue
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThinEngine{
				runtime:                tt.fields.runtime,
				name:                   tt.fields.name,
				namespace:              tt.fields.namespace,
				runtimeType:            tt.fields.runtimeType,
				Log:                    tt.fields.Log,
				Client:                 tt.fields.Client,
				gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				MetadataSyncDoneCh:     tt.fields.MetadataSyncDoneCh,
				runtimeInfo:            tt.fields.runtimeInfo,
				UnitTest:               tt.fields.UnitTest,
				retryShutdown:          tt.fields.retryShutdown,
				Helper:                 tt.fields.Helper,
			}
			t.parseWorkerImage(tt.args.runtime, tt.args.value)
		})
	}
}

func TestThinEngine_transform(t1 *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.ThinRuntime
		name                   string
		namespace              string
		runtimeType            string
		Log                    logr.Logger
		Client                 client.Client
		gracefulShutdownLimits int32
		MetadataSyncDoneCh     chan MetadataSyncResult
		runtimeInfo            base.RuntimeInfoInterface
		UnitTest               bool
		retryShutdown          int32
		Helper                 *ctrl.Helper
	}
	type args struct {
		runtime *datav1alpha1.ThinRuntime
		profile *datav1alpha1.ThinRuntimeProfile
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantValue *ThinValue
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThinEngine{
				runtime:                tt.fields.runtime,
				name:                   tt.fields.name,
				namespace:              tt.fields.namespace,
				runtimeType:            tt.fields.runtimeType,
				Log:                    tt.fields.Log,
				Client:                 tt.fields.Client,
				gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				MetadataSyncDoneCh:     tt.fields.MetadataSyncDoneCh,
				runtimeInfo:            tt.fields.runtimeInfo,
				UnitTest:               tt.fields.UnitTest,
				retryShutdown:          tt.fields.retryShutdown,
				Helper:                 tt.fields.Helper,
			}
			gotValue, err := t.transform(tt.args.runtime, tt.args.profile)
			if (err != nil) != tt.wantErr {
				t1.Errorf("transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t1.Errorf("transform() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestThinEngine_transformWorkers(t1 *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.ThinRuntime
		name                   string
		namespace              string
		runtimeType            string
		Log                    logr.Logger
		Client                 client.Client
		gracefulShutdownLimits int32
		MetadataSyncDoneCh     chan MetadataSyncResult
		runtimeInfo            base.RuntimeInfoInterface
		UnitTest               bool
		retryShutdown          int32
		Helper                 *ctrl.Helper
	}
	type args struct {
		runtime *datav1alpha1.ThinRuntime
		profile *datav1alpha1.ThinRuntimeProfile
		value   *ThinValue
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThinEngine{
				runtime:                tt.fields.runtime,
				name:                   tt.fields.name,
				namespace:              tt.fields.namespace,
				runtimeType:            tt.fields.runtimeType,
				Log:                    tt.fields.Log,
				Client:                 tt.fields.Client,
				gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				MetadataSyncDoneCh:     tt.fields.MetadataSyncDoneCh,
				runtimeInfo:            tt.fields.runtimeInfo,
				UnitTest:               tt.fields.UnitTest,
				retryShutdown:          tt.fields.retryShutdown,
				Helper:                 tt.fields.Helper,
			}
			if err := t.transformWorkers(tt.args.runtime, tt.args.profile, tt.args.value); (err != nil) != tt.wantErr {
				t1.Errorf("transformWorkers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
