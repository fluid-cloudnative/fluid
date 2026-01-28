/*
Copyright 2021 The Fluid Authors.
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
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newGooseEngineRT(c client.Client, name string, namespace string, withRuntimeInfo bool, unittest bool) *GooseFSEngine {
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.GooseFSRuntime)
	engine := &GooseFSEngine{
		runtime:     &datav1alpha1.GooseFSRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      c,
		runtimeInfo: nil,
		UnitTest:    unittest,
		Log:         fake.NullLogger(),
	}

	if withRuntimeInfo {
		engine.runtimeInfo = runTimeInfo
	}
	return engine
}

var _ = Describe("GetRuntimeInfo", func() {
	var fakeClient client.Client

	BeforeEach(func() {
		runtimeInputs := []*datav1alpha1.GooseFSRuntime{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Fuse: datav1alpha1.GooseFSFuseSpec{},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Fuse: datav1alpha1.GooseFSFuseSpec{},
				},
			},
		}
		daemonSetInputs := []*v1.DaemonSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-worker",
					Namespace: "fluid",
				},
				Spec: v1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hbase": "selector"}},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop-worker",
					Namespace: "fluid",
				},
				Spec: v1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hadoop": "selector"}},
					},
				},
			},
		}
		dataSetInputs := []*datav1alpha1.Dataset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop",
					Namespace: "fluid",
				},
			},
		}
		objs := []runtime.Object{}
		for _, runtimeInput := range runtimeInputs {
			objs = append(objs, runtimeInput.DeepCopy())
		}
		for _, daemonSetInput := range daemonSetInputs {
			objs = append(objs, daemonSetInput.DeepCopy())
		}
		for _, dataSetInput := range dataSetInputs {
			objs = append(objs, dataSetInput.DeepCopy())
		}
		fakeClient = fake.NewFakeClientWithScheme(testScheme, objs...)
	})

	type testCase struct {
		name            string
		namespace       string
		withRuntimeInfo bool
		unittest        bool
		expectErr       bool
		expectNil       bool
	}

	DescribeTable("should get runtime info correctly",
		func(tc testCase) {
			engine := newGooseEngineRT(fakeClient, tc.name, tc.namespace, tc.withRuntimeInfo, tc.unittest)
			runtimeInfo, err := engine.getRuntimeInfo()

			if tc.expectErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}

			if tc.expectNil {
				Expect(runtimeInfo).To(BeNil())
			} else {
				Expect(runtimeInfo).NotTo(BeNil())
			}
		},
		Entry("hbase without runtimeInfo, not unittest",
			testCase{
				name:            "hbase",
				namespace:       "fluid",
				withRuntimeInfo: false,
				unittest:        false,
				expectErr:       false,
				expectNil:       false,
			},
		),
		Entry("hbase without runtimeInfo, unittest",
			testCase{
				name:            "hbase",
				namespace:       "fluid",
				withRuntimeInfo: false,
				unittest:        true,
				expectErr:       false,
				expectNil:       false,
			},
		),
		Entry("hbase with runtimeInfo",
			testCase{
				name:            "hbase",
				namespace:       "fluid",
				withRuntimeInfo: true,
				unittest:        false,
				expectErr:       false,
				expectNil:       false,
			},
		),
		Entry("hadoop without runtimeInfo",
			testCase{
				name:            "hadoop",
				namespace:       "fluid",
				withRuntimeInfo: false,
				unittest:        false,
				expectErr:       false,
				expectNil:       false,
			},
		),
	)
})
