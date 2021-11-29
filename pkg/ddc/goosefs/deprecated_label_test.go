package goosefs

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func getTestGooseFSEngine(client client.Client, name string, namespace string) *GooseFSEngine {
	runTime := &datav1alpha1.GooseFSRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "goosefs", datav1alpha1.TieredStore{})
	engine := &GooseFSEngine{
		runtime:     runTime,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         log.NullLogger{},
	}
	return engine
}

func TestGooseFSEngine_GetDeprecatedCommonLabelname(t *testing.T) {
	testCases := []struct {
		name      string
		namespace string
		out       string
	}{
		{
			name:      "hbase",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-hbase",
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-hadoop",
		},
		{
			name:      "fluid",
			namespace: "test",
			out:       "data.fluid.io/storage-test-fluid",
		},
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme)
	for _, test := range testCases {
		engine := getTestGooseFSEngine(fakeClient, test.name, test.namespace)
		out := engine.getDeprecatedCommonLabelname()
		if out != test.out {
			t.Errorf("input parameter is %s-%s,expected %s, got %s", test.namespace, test.name, test.out, out)
		}
	}

}

func TestGooseFSEngine_HasDeprecatedCommonLabelname(t *testing.T) {

	// worker-name = e.name+"-worker"
	daemonSetWithSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-worker",
			Namespace: "fluid",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hbase": "selector"}},
			},
		},
	}
	daemonSetWithoutSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hadoop-worker",
			Namespace: "fluid",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hbase": "selector"}},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, daemonSetWithSelector)
	runtimeObjs = append(runtimeObjs, daemonSetWithoutSelector)
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, daemonSetWithSelector)
	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObjs...)

	testCases := []struct {
		name      string
		namespace string
		out       bool
		isErr     bool
	}{
		{
			name:      "hbase",
			namespace: "fluid",
			out:       true,
			isErr:     false,
		},
		{
			name:      "none",
			namespace: "fluid",
			out:       false,
			isErr:     false,
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			out:       false,
			isErr:     false,
		},
	}

	for _, test := range testCases {
		engine := getTestGooseFSEngine(fakeClient, test.name, test.namespace)
		out, err := engine.HasDeprecatedCommonLabelname()
		if out != test.out {
			t.Errorf("input parameter is %s-%s,expected %t, got %t", test.namespace, test.name, test.out, out)
		}
		isErr := err != nil
		if isErr != test.isErr {
			t.Errorf("input parameter is %s-%s,expected %t, got %t", test.namespace, test.name, test.isErr, isErr)
		}
	}
}
