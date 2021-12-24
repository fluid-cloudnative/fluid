package serverless

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"gopkg.in/yaml.v3"

	appsv1 "k8s.io/api/apps/v1"
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
	bTrue := true

	testcases := []testCase{
		{
			name: "inject_pod_success",
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						common.Serverless: common.True,
					},
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
					Image:   "test",
					SecurityContext: &corev1.SecurityContext{
						Privileged: &bTrue,
					}},
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
						Name: "fuse-device",
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
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						common.Serverless: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "fuse",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "test",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							},
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
					Volumes: []corev1.Volume{
						{Name: "dataset1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime_mnt/dataset1",
								},
							}}, {
							Name: "fuse-device",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		}, {
			name: "inject_deploy_success",
			in: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						common.Serverless: common.True,
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
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
				},
			},
			template: common.ServerlessInjectionTemplate{
				FuseContainer: corev1.Container{Name: "fuse",
					Args: []string{
						"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
					},
					Command: []string{"/entrypoint.sh"},
					Image:   "test",
					SecurityContext: &corev1.SecurityContext{
						Privileged: &bTrue,
					}},
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
						Name: "fuse-device",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/dev/fuse",
								Type: &hostPathCharDev,
							},
						},
					},
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						common.Serverless: common.True,
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "fuse",
								Args: []string{
									"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
								},
								Command: []string{"/entrypoint.sh"},
								Image:   "test",
								SecurityContext: &corev1.SecurityContext{
									Privileged: &bTrue,
								},
							},
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
								{Name: "dataset1",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/dataset1",
										},
									}}, {
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, testcase := range testcases {
		out, err := InjectObject(testcase.in, testcase.template)
		if testcase.wantErr == (err == nil) {
			t.Errorf("testcase %s failed, wantErr %v, Got error %v", testcase.name, testcase.wantErr, err)
		}

		if !reflect.DeepEqual(testcase.want, out) {
			want, err := yaml.Marshal(testcase.want)
			if err != nil {
				t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
			}

			outYaml, err := yaml.Marshal(out)
			if err != nil {
				t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
			}

			t.Errorf("testcase %s failed, want %v, Got  %v", testcase.name, string(want), string(outYaml))
		}

	}
}
