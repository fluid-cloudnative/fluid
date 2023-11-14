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

package efc

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime  *datav1alpha1.EFCRuntime
		efcValue *EFC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{},
		}, &EFC{}},
	}
	for _, test := range tests {
		engine := &EFCEngine{Log: fake.NullLogger()}
		err := engine.transformResourcesForWorker(test.runtime, test.efcValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if result, found := test.efcValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForWorkerWithValue(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("2Gi")
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")

	var tests = []struct {
		runtime                *datav1alpha1.EFCRuntime
		efcValue               *EFC
		wantedResourcesRequest string
		wantErr                bool
	}{
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Worker: datav1alpha1.EFCCompTemplateSpec{
						Resources: resources,
					},
				},
			},
			efcValue:               &EFC{},
			wantedResourcesRequest: "2Gi",
			wantErr:                false,
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test2",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Worker: datav1alpha1.EFCCompTemplateSpec{
						Resources: resources,
					},
				},
			},
			efcValue: &EFC{
				Worker: Worker{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								MediumType: string(common.Memory),
								Quota:      "3GB",
							},
						},
					},
				},
			},
			wantedResourcesRequest: "3Gi",
			wantErr:                false,
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test3",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Worker: datav1alpha1.EFCCompTemplateSpec{
						Resources: resources,
					},
				},
			},
			efcValue: &EFC{
				Worker: Worker{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								MediumType: string(common.Memory),
								Quota:      "5GB",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime)
		engine := &EFCEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformResourcesForWorker(test.runtime, test.efcValue)
		if (err == nil) != !test.wantErr {
			t.Error(err)
		}
		if err != nil {
			continue
		}
		quantity := test.efcValue.Worker.Resources.Requests[corev1.ResourceMemory]
		if quantity != test.wantedResourcesRequest {
			t.Errorf("expected %v, got %v", test.wantedResourcesRequest, test.efcValue.Worker.Resources.Requests[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForMasterNoValue(t *testing.T) {
	var tests = []struct {
		runtime  *datav1alpha1.EFCRuntime
		efcValue *EFC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{},
		}, &EFC{}},
	}
	for _, test := range tests {
		engine := &EFCEngine{Log: fake.NullLogger()}
		err := engine.transformResourcesForMaster(test.runtime, test.efcValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if result, found := test.efcValue.Master.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForMasterWithValue(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("2Gi")
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")

	var tests = []struct {
		runtime                *datav1alpha1.EFCRuntime
		efcValue               *EFC
		wantedResourcesRequest string
		wantErr                bool
	}{
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Master: datav1alpha1.EFCCompTemplateSpec{
						Resources: resources,
					},
				},
			},
			efcValue:               &EFC{},
			wantedResourcesRequest: "2Gi",
			wantErr:                false,
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test2",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Master: datav1alpha1.EFCCompTemplateSpec{
						Resources: resources,
					},
				},
			},
			efcValue: &EFC{
				Master: Master{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								MediumType: string(common.Memory),
								Quota:      "3GB",
							},
						},
					},
				},
			},
			wantedResourcesRequest: "3Gi",
			wantErr:                false,
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test3",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Master: datav1alpha1.EFCCompTemplateSpec{
						Resources: resources,
					},
				},
			},
			efcValue: &EFC{
				Master: Master{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								MediumType: string(common.Memory),
								Quota:      "5GB",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime)
		engine := &EFCEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformResourcesForMaster(test.runtime, test.efcValue)
		if (err == nil) != !test.wantErr {
			t.Error(err)
		}
		if err != nil {
			continue
		}
		quantity := test.efcValue.Master.Resources.Requests[corev1.ResourceMemory]
		if quantity != test.wantedResourcesRequest {
			t.Errorf("expected %v, got %v", test.wantedResourcesRequest, test.efcValue.Master.Resources.Requests[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime  *datav1alpha1.EFCRuntime
		efcValue *EFC
	}{
		{&datav1alpha1.EFCRuntime{
			Spec: datav1alpha1.EFCRuntimeSpec{},
		}, &EFC{}},
	}
	for _, test := range tests {
		engine := &EFCEngine{Log: fake.NullLogger()}
		err := engine.transformResourcesForFuse(test.runtime, test.efcValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if result, found := test.efcValue.Fuse.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForFuseWithValue(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("2Gi")
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")

	var tests = []struct {
		runtime                *datav1alpha1.EFCRuntime
		efcValue               *EFC
		wantedResourcesRequest string
		wantErr                bool
	}{
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Fuse: datav1alpha1.EFCFuseSpec{
						Resources: resources,
					},
				},
			},
			efcValue:               &EFC{},
			wantedResourcesRequest: "2Gi",
			wantErr:                false,
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test2",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Fuse: datav1alpha1.EFCFuseSpec{
						Resources: resources,
					},
				},
			},
			efcValue: &EFC{
				Fuse: Fuse{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								MediumType: string(common.Memory),
								Quota:      "3GB",
							},
						},
					},
				},
			},
			wantedResourcesRequest: "3Gi",
			wantErr:                false,
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test3",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					Fuse: datav1alpha1.EFCFuseSpec{
						Resources: resources,
					},
				},
			},
			efcValue: &EFC{
				Fuse: Fuse{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								MediumType: string(common.Memory),
								Quota:      "5GB",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime)
		engine := &EFCEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformResourcesForFuse(test.runtime, test.efcValue)
		if (err == nil) != !test.wantErr {
			t.Error(err)
		}
		if err != nil {
			continue
		}
		quantity := test.efcValue.Fuse.Resources.Requests[corev1.ResourceMemory]
		if quantity != test.wantedResourcesRequest {
			t.Errorf("expected %v, got %v", test.wantedResourcesRequest, test.efcValue.Fuse.Resources.Requests[corev1.ResourceMemory])
		}
	}
}
