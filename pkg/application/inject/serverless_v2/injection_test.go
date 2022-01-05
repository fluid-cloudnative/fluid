package serverless_v2

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	reflectutil "github.com/fluid-cloudnative/fluid/pkg/utils/reflect"
	"gopkg.in/yaml.v3"
	utilpointer "k8s.io/utils/pointer"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
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
					Name:       "test",
					Generation: 0,
					Annotations: map[string]string{
						common.Serverless: common.True,
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: utilpointer.Int32Ptr(0),
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

		gotContainers, gotVolumes, err := getInjectPiece(out)
		if err != nil {
			t.Errorf("testcase %s failed due to inject error %v", testcase.name, err)
		}

		wantContainers, wantVolumes, err := getInjectPiece(testcase.want)
		if err != nil {
			t.Errorf("testcase %s failed due to expect error %v", testcase.name, err)
		}

		if !reflect.DeepEqual(gotContainers, wantContainers) {
			want, err := yaml.Marshal(wantContainers)
			if err != nil {
				t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
			}

			outYaml, err := yaml.Marshal(gotContainers)
			if err != nil {
				t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
			}

			t.Errorf("testcase %s failed, want %v, Got  %v", testcase.name, string(want), string(outYaml))
		}

		if !reflect.DeepEqual(gotVolumes, wantVolumes) {
			want, err := yaml.Marshal(wantVolumes)
			if err != nil {
				t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
			}

			outYaml, err := yaml.Marshal(gotVolumes)
			if err != nil {
				t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
			}

			t.Errorf("testcase %s failed, want %v, Got  %v", testcase.name, string(want), string(outYaml))
		}

	}
}

func getInjectPiece(object runtime.Object) ([]corev1.Container, []corev1.Volume, error) {
	_, containersValue, err := reflectutil.ContainersValueFromObject(object, "", []string{"init"})
	if err != nil {
		return nil, nil, fmt.Errorf("get container references failed  with error %v", err)
	}

	_, volumesValue, err := reflectutil.VolumesValueFromObject(object, "", []string{})
	if err != nil {
		return nil, nil, fmt.Errorf("get volume Reference with error %v", err)
	}

	return containersValue.Interface().([]corev1.Container), volumesValue.Interface().([]corev1.Volume), nil

}

const inputYaml = `
apiVersion: "kubeflow.org/v1"
kind: "TFJob"
metadata:
  name: "mnist"
  namespace: kubeflow
  annotations:
   fluid.io/serverless: true
spec:
  cleanPodPolicy: None 
  tfReplicaSpecs:
    Worker:
      replicas: 1 
      restartPolicy: Never
      template:
        spec:
          containers:
            - name: tensorflow
              image: gcr.io/kubeflow-ci/tf-mnist-with-summaries:1.0
              command:
                - "python"
                - "/var/tf_mnist/mnist_with_summaries.py"
                - "--log_dir=/train/logs"
                - "--learning_rate=0.01"
                - "--batch_size=150"
              volumeMounts:
                - mountPath: "/train"
                  name: "training"
          volumes:
            - name: "training"
              persistentVolumeClaim:
                claimName: "tfevent-volume"  
`

func TestInjectObjectForUnstructed(t *testing.T) {

	obj := &unstructured.Unstructured{}
	hostPathCharDev := corev1.HostPathCharDev
	bTrue := true

	dec := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode([]byte(inputYaml), nil, obj)
	if err != nil {
		t.Errorf("Failed to decode due to %v", err)
	}

	// Get the common metadata, and show GVK
	fmt.Println(obj.GetName(), gvk.String())

	template := common.ServerlessInjectionTemplate{
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
	}

	out, err := InjectObject(obj, template)
	if err != nil {
		t.Errorf("Failed to InjectObject due to %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	enc.Encode(out)

}
