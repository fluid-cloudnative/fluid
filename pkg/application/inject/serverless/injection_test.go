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

	hostPathCharDev := corev1.HostPathCharDev

	testcases := []testCase{
		{
			name: "inject_success",
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset1",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset1",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset1",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			template: common.ServerlessInjectionTemplate{
				FuseContainer: corev1.Container{Name: "fuse",
					Args: []string{
						"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
					},
					Command: []string{"/entrypoint.sh"},
					Image:   "test"},
				VolumesToUpdate: []corev1.Volume{
					{
						Name: "dataset1",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/runtime_mnt/dataset1",
							},
						},
					},
				},
				VolumesToAdd: []corev1.Volume{
					{
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/dev/fuse",
								Type: &hostPathCharDev,
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "fuse",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "test",
						}, {
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset1",
									MountPath: "/data",
								},
							},
						},
					},
				},
			},
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
