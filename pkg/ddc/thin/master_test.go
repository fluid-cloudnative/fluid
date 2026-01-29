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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Master Tests", func() {
	Describe("ThinEngine.CheckMasterReady", func() {
		Context("when fuse daemonset exists", func() {
			It("should return ready", func() {
				daemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-fuse",
						Namespace: "fluid",
					},
				}
				testObjs := []runtime.Object{daemonSet.DeepCopy()}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:      "test",
					namespace: "fluid",
					Client:    client,
				}

				ready, err := engine.CheckMasterReady()
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})

		Context("when fuse daemonset does not exist", func() {
			It("should return error", func() {
				testObjs := []runtime.Object{}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:      "notexist",
					namespace: "fluid",
					Client:    client,
				}

				ready, err := engine.CheckMasterReady()
				Expect(err).NotTo(BeNil())
				Expect(ready).To(BeFalse())
			})
		})
	})

	Describe("ThinEngine.ShouldSetupMaster", func() {
		DescribeTable("should correctly determine if master setup is needed",
			func(phase datav1alpha1.RuntimePhase, expected bool) {
				runtimeInput := &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-" + string(phase),
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{
						FusePhase: phase,
					},
				}
				testObjs := []runtime.Object{runtimeInput.DeepCopy()}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:      "test-" + string(phase),
					namespace: "fluid",
					Client:    client,
				}

				should, err := engine.ShouldSetupMaster()
				Expect(err).To(BeNil())
				Expect(should).To(Equal(expected))
			},
			Entry("None phase - should setup", datav1alpha1.RuntimePhaseNone, true),
			Entry("Ready phase - should not setup", datav1alpha1.RuntimePhaseReady, false),
			Entry("NotReady phase - should not setup", datav1alpha1.RuntimePhaseNotReady, false),
		)
	})

	Describe("ThinEngine.SetupMaster", func() {
		Context("when fuse daemonset already exists", func() {
			It("should return no error", func() {
				daemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "existing-fuse",
						Namespace: "fluid",
					},
				}
				testObjs := []runtime.Object{daemonSet.DeepCopy()}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:      "existing",
					namespace: "fluid",
					Client:    client,
					Log:       fake.NullLogger(),
				}

				err := engine.SetupMaster()
				Expect(err).To(BeNil())
			})
		})
	})
})
