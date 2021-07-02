package alluxio

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)



func getTestAlluxioEngine(client client.Client, name string, namespace string) *AlluxioEngine {
	runTime := &datav1alpha1.AlluxioRuntime{}
	runTimeInfo,_ := base.BuildRuntimeInfo(name,namespace,"alluxio",datav1alpha1.Tieredstore{})
	engine := &AlluxioEngine{
		runtime:                runTime,
		name:                   name,
		namespace:              namespace,
		Client:                 client,
		runtimeInfo:            runTimeInfo,
		Log:                    NullLogger{},
	}
	return engine
}



func TestAlluxioEngine_GetDeprecatedCommonLabelname(t *testing.T){

}

func TestAlluxioEngine_HasDeprecatedCommonLabelname(t *testing.T){

	// worker-name = e.name+"-worker"
	daemonset := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
		Name: "hbase-worker",
		Namespace: "fluid",
	},
		Spec: v1.DaemonSetSpec{
			Template:corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hbase":"selector"}},
			},
		},

	}
	runtimeObjs  := []runtime.Object{}
	runtimeObjs  = append(runtimeObjs, daemonset)
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion,daemonset)
	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObjs...)
	alluxioEngine := getTestAlluxioEngine(fakeClient,"hbase","fluid")
	alluxioEngine.HasDeprecatedCommonLabelname()

}

