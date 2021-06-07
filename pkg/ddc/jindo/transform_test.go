/*

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

package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestTransformTolerations(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     int
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				Tieredstore: datav1alpha1.Tieredstore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
				Master: datav1alpha1.JindoCompTemplateSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "master",
						Operator: "Equals",
						Value:    "true",
					}},
				},
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "worker",
						Operator: "Equals",
						Value:    "true",
					}},
				},
				Fuse: datav1alpha1.JindoFuseSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "fuse",
						Operator: "Equals",
						Value:    "true",
					}},
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
				Tolerations: []corev1.Toleration{{
					Key:      "jindo",
					Operator: "Equals",
					Value:    "true",
				}},
			}}, &Jindo{}, 2,
		},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: log.NullLogger{}}
		engine.transformTolerations(test.dataset, test.runtime, test.jindoValue)
		if len(test.jindoValue.Master.Tolerations) != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Master.Tolerations)
		}
		if len(test.jindoValue.Worker.Tolerations) != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Master.Tolerations)
		}
		if len(test.jindoValue.Fuse.Tolerations) != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Master.Tolerations)
		}
	}
}
