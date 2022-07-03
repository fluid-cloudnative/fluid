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

package jindofsx

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestJindoFSxEngine_syncMasterSpec(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
		// runtimeType            string
		Log    logr.Logger
		Client client.Client
		// gracefulShutdownLimits int32
		// retryShutdown          int32
		// runtimeInfo            base.RuntimeInfoInterface
		MetadataSyncDoneCh chan MetadataSyncResult
		// cacheNodeNames         []string
		Recorder record.EventRecorder
		Helper   *ctrl.Helper
	}
	type args struct {
		ctx     cruntime.ReconcileRequestContext
		runtime *datav1alpha1.JindoRuntime
		master  appsv1.StatefulSet
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantChanged bool
		wantErr     bool
	}{
		{
			name:   "Not resource for jindoruntime",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.args.master.DeepCopy())

			s := runtime.NewScheme()
			data := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.fields.name,
					Namespace: tt.fields.namespace,
				},
			}
			s.AddKnownTypes(appsv1.SchemeGroupVersion, &tt.args.master)

			_ = v1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.args.runtime)
			runtimeObjs = append(runtimeObjs, data)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoFSxEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncMasterSpec(tt.args.ctx, tt.args.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("JindoFSxEngine.syncMasterSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("JindoFSxEngine.syncMasterSpec() = %v, want %v", gotChanged, tt.wantChanged)
			}
		})
	}
}
