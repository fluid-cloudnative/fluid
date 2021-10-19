package juicefs

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.JuiceFSRuntime
		juicefsValue *JuiceFS
	}{
		{&datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{},
		}, &JuiceFS{}},
	}
	for _, test := range tests {
		engine := &JuiceFSEngine{Log: log.NullLogger{}}
		engine.transformResourcesForWorker(test.runtime, test.juicefsValue)
		if result, found := test.juicefsValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.JuiceFSRuntime
		juicefsValue *JuiceFS
	}{
		{&datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{},
		}, &JuiceFS{}},
	}
	for _, test := range tests {
		engine := &JuiceFSEngine{Log: log.NullLogger{}}
		engine.transformResourcesForFuse(test.runtime, test.juicefsValue)
		if result, found := test.juicefsValue.Fuse.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}
