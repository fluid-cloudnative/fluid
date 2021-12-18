package serverless

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestInjectObject(t *testing.T) {
	type testCase struct {
		name     string
		in       runtime.Object
		template common.ServerlessInjectionTemplate
		want     runtime.Object
		wantErr  bool
	}

	testcases := []testCase{
		{
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			template: common.ServerlessInjectionTemplate{
				FuseContainer: corev1.Container{},
			},
			want:    &corev1.Pod{},
			wantErr: true,
		},
	}

	for _, testcase := range testcases {
		out, err := InjectObject(testcase.in, testcase.template)
		if testcase.wantErr == (err == nil) {
			t.Errorf("testcase %s failed, wantErr %v, Got error %v", testcase.name, testcase.wantErr, err)
		}

		if !reflect.DeepEqual(testcase.want, out) {
			t.Errorf("testcase %s failed, want %v, Got  %v", testcase.name, testcase.want, out)
		}

	}
}
