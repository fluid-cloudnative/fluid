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
		runtime  *datav1alpha1.EACRuntime
		eacValue *EAC
	}{
		{&datav1alpha1.EACRuntime{
			Spec: datav1alpha1.EACRuntimeSpec{
				Master: datav1alpha1.EACCompTemplateSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EAC{},
		},
		{&datav1alpha1.EACRuntime{
			Spec: datav1alpha1.EACRuntimeSpec{
				Master: datav1alpha1.EACCompTemplateSpec{
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
		runtime  *datav1alpha1.EACRuntime
		eacValue *EAC
	}{
		{&datav1alpha1.EACRuntime{
			Spec: datav1alpha1.EACRuntimeSpec{
				Fuse: datav1alpha1.EACFuseSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EAC{},
		},
		{&datav1alpha1.EACRuntime{
			Spec: datav1alpha1.EACRuntimeSpec{
				Fuse: datav1alpha1.EACFuseSpec{
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
		runtime  *datav1alpha1.EACRuntime
		eacValue *EAC
	}{
		{&datav1alpha1.EACRuntime{
			Spec: datav1alpha1.EACRuntimeSpec{
				Worker: datav1alpha1.EACCompTemplateSpec{
					NetworkMode: "HostNetwork",
				},
			},
		},
			&EAC{},
		},
		{&datav1alpha1.EACRuntime{
			Spec: datav1alpha1.EACRuntimeSpec{
				Worker: datav1alpha1.EACCompTemplateSpec{
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
