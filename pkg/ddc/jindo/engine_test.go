package jindo

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestBuild(t *testing.T) {
	var namespace = v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, namespace.DeepCopy())

	var dataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	testObjs = append(testObjs, dataset.DeepCopy())

	var runtime = datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.JindoRuntimeSpec{
			Master: datav1alpha1.JindoCompTemplateSpec{
				Replicas: 1,
			},
			Fuse: datav1alpha1.JindoFuseSpec{
				Global: false,
			},
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}
	testObjs = append(testObjs, runtime.DeepCopy())

	var daemonset = appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-worker",
			Namespace: "fluid",
		},
	}
	testObjs = append(testObjs, daemonset.DeepCopy())
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	var ctx = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         log.NullLogger{},
		RuntimeType: common.JINDO_RUNTIME,
		Runtime:     &runtime,
	}

	engine, err := Build("testId", ctx)
	if err != nil || engine == nil {
		t.Errorf("fail to exec the build function with the eror %v", err)
	}

}
