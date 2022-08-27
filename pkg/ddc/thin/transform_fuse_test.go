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
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestThinEngine_parseFromProfileFuse(t1 *testing.T) {
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
			if err := t.parseFromProfileFuse(tt.args.profile, tt.args.value); (err != nil) != tt.wantErr {
				t1.Errorf("parseFromProfileFuse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThinEngine_parseFuseImage(t1 *testing.T) {
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
			t.parseFuseImage(tt.args.runtime, tt.args.value)
		})
	}
}

func TestThinEngine_parseFuseOptions(t1 *testing.T) {
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
		dataset *datav1alpha1.Dataset
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOption string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := ThinEngine{
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
			if gotOption := t.parseFuseOptions(tt.args.runtime, tt.args.profile, tt.args.dataset); gotOption != tt.wantOption {
				t1.Errorf("parseFuseOptions() = %v, want %v", gotOption, tt.wantOption)
			}
		})
	}
}

func TestThinEngine_transformFuse(t1 *testing.T) {
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
		dataset *datav1alpha1.Dataset
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
			if err := t.transformFuse(tt.args.runtime, tt.args.profile, tt.args.dataset, tt.args.value); (err != nil) != tt.wantErr {
				t1.Errorf("transformFuse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
