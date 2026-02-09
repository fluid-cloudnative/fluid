/*
Copyright 2026 The Fluid Authors.

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

package watch

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("DatasetEventHandler", func() {
	var handler *datasetEventHandler

	BeforeEach(func() {
		handler = &datasetEventHandler{}
	})

	Describe("Instantiation", func() {
		It("should be instantiable", func() {
			Expect(handler).NotTo(BeNil())
		})
	})

	Describe("onUpdateFunc", func() {
		It("should return a non-nil function", func() {
			updateFunc := handler.onUpdateFunc("alluxio")
			Expect(updateFunc).NotTo(BeNil())
		})

		Context("when handling update events", func() {
			type testCase struct {
				runtimeType  string
				event        event.UpdateEvent
				expectUpdate bool
				description  string
			}

			DescribeTable("should handle various update scenarios",
				func(tc testCase) {
					updateFunc := handler.onUpdateFunc(tc.runtimeType)
					result := updateFunc(tc.event)
					Expect(result).To(Equal(tc.expectUpdate), tc.description)
				},
				Entry("ObjectNew is not a Dataset",
					testCase{
						runtimeType: "alluxio",
						event: event.UpdateEvent{
							ObjectOld: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "1",
								},
							},
							ObjectNew: &corev1.Pod{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-pod",
									Namespace:       "default",
									ResourceVersion: "2",
								},
							},
						},
						expectUpdate: false,
						description:  "Should return false when ObjectNew is not a Dataset",
					},
				),
				Entry("Runtime type does not match",
					testCase{
						runtimeType: "alluxio",
						event: event.UpdateEvent{
							ObjectOld: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "1",
								},
							},
							ObjectNew: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "2",
								},
								Status: datav1alpha1.DatasetStatus{
									Runtimes: []datav1alpha1.Runtime{
										{
											Type: "jindo",
										},
									},
								},
							},
						},
						expectUpdate: false,
						description:  "Should return false when runtime type doesn't match",
					},
				),

				Entry("ObjectOld is not a Dataset",
					testCase{
						runtimeType: "alluxio",
						event: event.UpdateEvent{
							ObjectOld: &corev1.Pod{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-pod",
									Namespace:       "default",
									ResourceVersion: "1",
								},
							},
							ObjectNew: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "2",
								},
								Status: datav1alpha1.DatasetStatus{
									Runtimes: []datav1alpha1.Runtime{
										{
											Type: "alluxio",
										},
									},
								},
							},
						},
						expectUpdate: false,
						description:  "Should return false when ObjectOld is not a Dataset",
					},
				),
				Entry("ResourceVersion is the same",
					testCase{
						runtimeType: "alluxio",
						event: event.UpdateEvent{
							ObjectOld: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "1",
								},
							},
							ObjectNew: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "1",
								},
								Status: datav1alpha1.DatasetStatus{
									Runtimes: []datav1alpha1.Runtime{
										{
											Type: "alluxio",
										},
									},
								},
							},
						},
						expectUpdate: false,
						description:  "Should return false when ResourceVersion hasn't changed",
					},
				),
				Entry("Valid update with matching runtime type",
					testCase{
						runtimeType: "alluxio",
						event: event.UpdateEvent{
							ObjectOld: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "1",
								},
							},
							ObjectNew: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "2",
								},
								Status: datav1alpha1.DatasetStatus{
									Runtimes: []datav1alpha1.Runtime{
										{
											Type: "alluxio",
										},
									},
								},
							},
						},
						expectUpdate: true,
						description:  "Should return true when all conditions are met",
					},
				),
				Entry("Valid update with empty runtimes",
					testCase{
						runtimeType: "alluxio",
						event: event.UpdateEvent{
							ObjectOld: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "1",
								},
							},
							ObjectNew: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "2",
								},
								Status: datav1alpha1.DatasetStatus{
									Runtimes: []datav1alpha1.Runtime{},
								},
							},
						},
						expectUpdate: true,
						description:  "Should return true when runtimes slice is empty",
					},
				),
				Entry("Valid update with no runtimes status",
					testCase{
						runtimeType: "jindo",
						event: event.UpdateEvent{
							ObjectOld: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "1",
								},
							},
							ObjectNew: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "2",
								},
								Status: datav1alpha1.DatasetStatus{},
							},
						},
						expectUpdate: true,
						description:  "Should return true when Status.Runtimes is nil",
					},
				),
				Entry("Valid update with different runtime type (jindo)",
					testCase{
						runtimeType: "jindo",
						event: event.UpdateEvent{
							ObjectOld: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "100",
								},
							},
							ObjectNew: &datav1alpha1.Dataset{
								ObjectMeta: metav1.ObjectMeta{
									Name:            "test-dataset",
									Namespace:       "default",
									ResourceVersion: "101",
								},
								Status: datav1alpha1.DatasetStatus{
									Runtimes: []datav1alpha1.Runtime{
										{
											Type: "jindo",
										},
									},
								},
							},
						},
						expectUpdate: true,
						description:  "Should return true for jindo runtime type",
					},
				),
			)
		})
	})
})
