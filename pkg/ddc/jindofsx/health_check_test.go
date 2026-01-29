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

package jindofsx

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("JindoFSxEngine Health Check Tests", Label("pkg.ddc.jindofsx.health_check_test.go"), func() {
	var (
		jindoRuntime *datav1alpha1.JindoRuntime
		dataset      *datav1alpha1.Dataset
		worker       *appsv1.StatefulSet
		master       *appsv1.StatefulSet
		fuse         *appsv1.DaemonSet
		engine       *JindoFSxEngine
		runtimeObjs  []runtime.Object
		client       client.Client
		name         string
		namespace    string
		runtimeInfo  base.RuntimeInfoInterface
	)

	BeforeEach(func() {
		name = "test-data"
		namespace = "big-data"

		jindoRuntime = &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 1,
				Fuse:     datav1alpha1.JindoFuseSpec{},
			},
		}

		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}

		worker = &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name + "-jindofs-worker",
				Namespace: namespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		}

		master = &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name + "-jindofs-master",
				Namespace: namespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		}

		fuse = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name + "-jindofs-fuse",
				Namespace: namespace,
			},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable: 0,
			},
		}

		runtimeObjs = []runtime.Object{jindoRuntime, dataset, worker, master, fuse}
	})

	JustBeforeEach(func() {
		s := runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
		_ = v1.AddToScheme(s)
		_ = appsv1.AddToScheme(s)

		client = fake.NewFakeClientWithScheme(s, runtimeObjs...)

		var err error
		runtimeInfo, err = base.BuildRuntimeInfo(name, namespace, common.JindoRuntime)
		Expect(err).To(BeNil())

		engine = &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      name,
			namespace: namespace,
			Client:    client,
			Log:       ctrl.Log.WithName(name),
			Recorder:  record.NewFakeRecorder(300),
		}

		engine.Helper = ctrlhelper.BuildHelper(runtimeInfo, client, engine.Log)
	})

	Describe("Test CheckRuntimeHealthy", func() {
		Context("when all components are healthy", func() {
			It("should not return error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).To(BeNil())
			})
		})

		Context("when master is not healthy", func() {
			BeforeEach(func() {
				master = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-master",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				}
				worker = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-worker",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
						Replicas:      1,
					},
				}
				fuse = &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-fuse",
						Namespace: namespace,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				}
				runtimeObjs = []runtime.Object{jindoRuntime, dataset, worker, master, fuse}
			})

			It("should return error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when worker is not healthy", func() {
			BeforeEach(func() {
				worker = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-worker",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
						Replicas:      1,
					},
				}
				master = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-master",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}
				fuse = &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-fuse",
						Namespace: namespace,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				}
				runtimeObjs = []runtime.Object{jindoRuntime, dataset, worker, master, fuse}
			})

			It("should return error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when fuse is not healthy but fuse is configured", func() {
			BeforeEach(func() {
				worker = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-worker",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}
				master = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-master",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}
				fuse = &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-fuse",
						Namespace: namespace,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				}
				runtimeObjs = []runtime.Object{jindoRuntime, dataset, worker, master, fuse}
			})

			It("should not return error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).To(BeNil())
			})
		})

		Context("when master not found", func() {
			BeforeEach(func() {
				jindoRuntime = &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-master",
						Namespace: namespace,
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				}
				name = "unhealthy-no-master"

				worker = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-worker",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}
				master = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}
				fuse = &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-no-master",
						Namespace: namespace,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				}
				runtimeObjs = []runtime.Object{jindoRuntime, dataset, worker, master, fuse}
			})

			It("should return error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when worker not found", func() {
			BeforeEach(func() {
				jindoRuntime = &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-worker",
						Namespace: namespace,
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				}
				name = "unhealthy-no-worker"

				worker = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-master",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}
				master = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs",
						Namespace: namespace,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}
				fuse = &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-jindofs-no-worker",
						Namespace: namespace,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				}
				runtimeObjs = []runtime.Object{jindoRuntime, dataset, worker, master, fuse}
			})

			It("should return error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
