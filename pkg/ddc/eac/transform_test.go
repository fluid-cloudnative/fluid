package eac

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestEACEngine_transform(t *testing.T) {
	var tests = []struct {
		runtime *datav1alpha1.EACRuntime
		dataset *datav1alpha1.Dataset
	}{
		{&datav1alpha1.EACRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EACRuntimeSpec{
				Fuse: datav1alpha1.EACFuseSpec{},
				Worker: datav1alpha1.EACCompTemplateSpec{
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
						MountPoint: "eac://abcd-abc67.cn-zhangjiakou.nas.aliyuncs.com:/test-fluid-3/",
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
		engine := EACEngine{
			name:      "test",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
			runtime:   test.runtime,
		}
		ctrl.SetLogger(zap.New(func(o *zap.Options) {
			o.Development = true
		}))
		err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
		if err != nil {
			t.Fatal(err.Error())
		}
		_, err = engine.transform(test.runtime)
		if err != nil {
			t.Errorf("error %v", err)
		}
	}
}
