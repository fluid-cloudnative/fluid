/*
Copyright 2023 The Fluid Author.

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
package eac

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/util/net"
)

func TestTransformPortForMaster(t *testing.T) {
	var tests = []struct {
		runtime  *datav1alpha1.EFCRuntime
		eacValue *EAC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EAC{},
		},
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					NetworkMode: "ContainerNetwork",
				},
			},
		},
			&EAC{},
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime.DeepCopy())
		engine := &EACEngine{Log: fake.NullLogger()}
		err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatalf("failed to set up runtime port allocator due to %v", err)
		}
		err = engine.transformPortForMaster(test.runtime, test.eacValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
	}
}

func TestTransformPortForFuse(t *testing.T) {
	var tests = []struct {
		runtime  *datav1alpha1.EFCRuntime
		eacValue *EAC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EAC{},
		},
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{
					NetworkMode: "ContainerNetwork",
				},
			},
		},
			&EAC{},
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime.DeepCopy())
		engine := &EACEngine{Log: fake.NullLogger()}
		err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatal(err.Error())
		}
		err = engine.transformPortForFuse(test.runtime, test.eacValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
	}
}

func TestTransformPortForWorker(t *testing.T) {
	var tests = []struct {
		runtime  *datav1alpha1.EFCRuntime
		eacValue *EAC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Worker: datav1alpha1.EFCCompTemplateSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EAC{},
		},
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Worker: datav1alpha1.EFCCompTemplateSpec{
					NetworkMode: "ContainerNetwork",
				},
			},
		},
			&EAC{},
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime.DeepCopy())
		engine := &EACEngine{Log: fake.NullLogger()}
		err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatal(err.Error())
		}
		err = engine.transformPortForWorker(test.runtime, test.eacValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
	}
}
