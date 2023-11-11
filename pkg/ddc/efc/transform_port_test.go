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
package efc

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
		efcValue *EFC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EFC{},
		},
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					NetworkMode: "ContainerNetwork",
				},
			},
		},
			&EFC{},
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime.DeepCopy())
		engine := &EFCEngine{Log: fake.NullLogger()}
		err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatalf("failed to set up runtime port allocator due to %v", err)
		}
		err = engine.transformPortForMaster(test.runtime, test.efcValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
	}
}

func TestTransformPortForFuse(t *testing.T) {
	var tests = []struct {
		runtime  *datav1alpha1.EFCRuntime
		efcValue *EFC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EFC{},
		},
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{
					NetworkMode: "ContainerNetwork",
				},
			},
		},
			&EFC{},
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime.DeepCopy())
		engine := &EFCEngine{Log: fake.NullLogger()}
		err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatal(err.Error())
		}
		err = engine.transformPortForFuse(test.runtime, test.efcValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
	}
}

func TestTransformPortForWorker(t *testing.T) {
	var tests = []struct {
		runtime  *datav1alpha1.EFCRuntime
		efcValue *EFC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Worker: datav1alpha1.EFCCompTemplateSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EFC{},
		},
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{
				Worker: datav1alpha1.EFCCompTemplateSpec{
					NetworkMode: "ContainerNetwork",
				},
			},
		},
			&EFC{},
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime.DeepCopy())
		engine := &EFCEngine{Log: fake.NullLogger()}
		err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatal(err.Error())
		}
		err = engine.transformPortForWorker(test.runtime, test.efcValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
	}
}
