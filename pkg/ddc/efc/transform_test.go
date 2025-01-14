/*
Copyright 2023 The Fluid Authors.

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

package efc

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestEFCEngine_transform(t *testing.T) {
	var tests = []struct {
		runtime *datav1alpha1.EFCRuntime
		dataset *datav1alpha1.Dataset
	}{
		{&datav1alpha1.EFCRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{},
				Worker: datav1alpha1.EFCCompTemplateSpec{
					Replicas: 2,
				},
			},
		}, &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "nfs://abcd-abc67.cn-zhangjiakou.nas.aliyuncs.com:/test-fluid-3/",
					},
				},
			},
		},
		},
	}
	for _, test := range tests {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, test.runtime.DeepCopy())
		testObjs = append(testObjs, test.dataset.DeepCopy())

		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
		runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "efc")
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		engine := EFCEngine{
			name:        "test",
			namespace:   "fluid",
			Client:      client,
			Log:         fake.NullLogger(),
			runtime:     test.runtime,
			runtimeInfo: runtimeInfo,
		}
		ctrl.SetLogger(zap.New(func(o *zap.Options) {
			o.Development = true
		}))
		err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatal(err.Error())
		}
		_, err = engine.transform(test.runtime)
		if err != nil {
			t.Errorf("error %v", err)
		}
	}
}
