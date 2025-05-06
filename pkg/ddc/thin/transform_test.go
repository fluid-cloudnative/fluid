/*
  Copyright 2022 The Fluid Authors.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package thin

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestThinEngine_transformTolerations(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	type args struct {
		dataset *datav1alpha1.Dataset
		value   *ThinValue
	}
	var tests = []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test",
			fields: fields{
				name:      "",
				namespace: "",
			},
			args: args{
				dataset: &datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "a",
						Operator: corev1.TolerationOpEqual,
						Value:    "b",
					}},
				}},
				value: &ThinValue{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &ThinEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			j.transformTolerations(tt.args.dataset, tt.args.value)
			if len(tt.args.value.Tolerations) != len(tt.args.dataset.Spec.Tolerations) {
				t.Errorf("transformTolerations() tolerations = %v", tt.args.value.Tolerations)
			}
		})
	}
}

func TestThinEngine_parseFromProfile(t1 *testing.T) {
	profile := datav1alpha1.ThinRuntimeProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.ThinRuntimeProfileSpec{
			Worker: datav1alpha1.ThinCompTemplateSpec{
				Image:           "test",
				ImageTag:        "v1",
				ImagePullPolicy: "Always",
				Env: []corev1.EnvVar{{
					Name:  "a",
					Value: "b",
				}, {
					Name: "b",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-cm",
							},
						},
					},
				}},
				NodeSelector: map[string]string{"a": "b"},
				Ports: []corev1.ContainerPort{{
					Name:          "port",
					ContainerPort: 8080,
				}},
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/healthz",
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/healthz",
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				NetworkMode: datav1alpha1.HostNetworkMode,
			},
		},
	}
	wantValue := &ThinValue{
		Worker: Worker{
			Image:           "test",
			ImageTag:        "v1",
			ImagePullPolicy: "Always",
			Resources: common.Resources{
				Requests: map[corev1.ResourceName]string{},
				Limits:   map[corev1.ResourceName]string{},
			},
			HostNetwork: true,
			Envs: []corev1.EnvVar{{
				Name:  "a",
				Value: "b",
			}, {
				Name: "b",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-cm",
						},
					},
				},
			}},
			NodeSelector: map[string]string{"a": "b"},
			Ports: []corev1.ContainerPort{{
				Name:          "port",
				ContainerPort: 8080,
			}},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/healthz",
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      1,
				PeriodSeconds:       1,
				SuccessThreshold:    1,
				FailureThreshold:    1,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/healthz",
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      1,
				PeriodSeconds:       1,
				SuccessThreshold:    1,
				FailureThreshold:    1,
			},
		},
	}
	value := &ThinValue{}
	t1.Run("test", func(t1 *testing.T) {
		t := &ThinEngine{
			Log: fake.NullLogger(),
		}
		t.parseFromProfile(&profile, value)
		if !reflect.DeepEqual(value, wantValue) {
			t1.Errorf("parseFromProfile() got = %v, want = %v", value, wantValue)
		}
	})
}

func TestThinEngine_parseWorkerImage(t1 *testing.T) {
	type args struct {
		runtime *datav1alpha1.ThinRuntime
		value   *ThinValue
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				runtime: &datav1alpha1.ThinRuntime{
					Spec: datav1alpha1.ThinRuntimeSpec{
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Image:           "test",
							ImageTag:        "v1",
							ImagePullPolicy: "Always",
						},
					},
				},
				value: &ThinValue{},
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThinEngine{}
			t.parseWorkerImage(tt.args.runtime, tt.args.value)
			if tt.args.value.Worker.Image != tt.args.runtime.Spec.Worker.Image ||
				tt.args.value.Worker.ImageTag != tt.args.runtime.Spec.Worker.ImageTag ||
				tt.args.value.Worker.ImagePullPolicy != tt.args.runtime.Spec.Worker.ImagePullPolicy {
				t1.Errorf("got %v, want %v", tt.args.value.Worker, tt.args.runtime.Spec.Worker)
			}
		})
	}
}
