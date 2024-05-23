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

package dataflowaffinity

import (
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestDataOpJobReconciler_injectPodNodeLabelsToJob(t *testing.T) {
	type fields struct {
		Client   client.Client
		Recorder record.EventRecorder
		Log      logr.Logger
	}
	type args struct {
		job *batchv1.Job
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
		t.Run(tt.name, func(t *testing.T) {
			f := &DataOpJobReconciler{
				Client:   tt.fields.Client,
				Recorder: tt.fields.Recorder,
				Log:      tt.fields.Log,
			}
			if err := f.injectPodNodeLabelsToJob(tt.args.job); (err != nil) != tt.wantErr {
				t.Errorf("injectPodNodeLabelsToJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
