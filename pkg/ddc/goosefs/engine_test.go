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

package goosefs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Build", func() {
	It("should build engine successfully", func() {
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

		var goosefsRuntime = datav1alpha1.GooseFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Master: datav1alpha1.GooseFSCompTemplateSpec{
					Replicas: 1,
				},
				Fuse: datav1alpha1.GooseFSFuseSpec{},
			},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		}
		testObjs = append(testObjs, goosefsRuntime.DeepCopy())

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
			Log:         fake.NullLogger(),
			RuntimeType: "goosefs",
			Runtime:     &goosefsRuntime,
		}

		engine, err := Build("testId", ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(engine).NotTo(BeNil())
	})
})
