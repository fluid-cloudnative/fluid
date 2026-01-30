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
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ThinEngine_ShouldSetupWorkers", func() {
	It("should return true when worker phase is None and worker is enabled", func() {
		runtime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test0",
				Namespace: "thin",
			},
			Spec: datav1alpha1.ThinRuntimeSpec{
				Worker: datav1alpha1.ThinCompTemplateSpec{
					Enabled: true,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				WorkerPhase: datav1alpha1.RuntimePhaseNone,
			},
		}

		data := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test0",
				Namespace: "thin",
			},
		}

		s := apimachineryruntime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, runtime, data)
		_ = v1.AddToScheme(s)

		mockClient := fake.NewFakeClientWithScheme(s, runtime, data)

		e := &ThinEngine{
			name:      "test0",
			namespace: "thin",
			runtime:   runtime,
			Client:    mockClient,
		}

		gotShould, err := e.ShouldSetupWorkers()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotShould).To(BeTrue())
	})

	It("should return false when worker phase is NotReady", func() {
		runtime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "thin",
			},
			Spec: datav1alpha1.ThinRuntimeSpec{
				Worker: datav1alpha1.ThinCompTemplateSpec{
					Enabled: true,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				WorkerPhase: datav1alpha1.RuntimePhaseNotReady,
			},
		}

		data := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "thin",
			},
		}

		s := apimachineryruntime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, runtime, data)
		_ = v1.AddToScheme(s)

		mockClient := fake.NewFakeClientWithScheme(s, runtime, data)

		e := &ThinEngine{
			name:      "test1",
			namespace: "thin",
			runtime:   runtime,
			Client:    mockClient,
		}

		gotShould, err := e.ShouldSetupWorkers()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotShould).To(BeFalse())
	})

	It("should return false when worker phase is PartialReady", func() {
		runtime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test2",
				Namespace: "thin",
			},
			Spec: datav1alpha1.ThinRuntimeSpec{
				Worker: datav1alpha1.ThinCompTemplateSpec{
					Enabled: true,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				WorkerPhase: datav1alpha1.RuntimePhasePartialReady,
			},
		}

		data := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test2",
				Namespace: "thin",
			},
		}

		s := apimachineryruntime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, runtime, data)
		_ = v1.AddToScheme(s)

		mockClient := fake.NewFakeClientWithScheme(s, runtime, data)

		e := &ThinEngine{
			name:      "test2",
			namespace: "thin",
			runtime:   runtime,
			Client:    mockClient,
		}

		gotShould, err := e.ShouldSetupWorkers()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotShould).To(BeFalse())
	})

	It("should return false when worker phase is Ready", func() {
		runtime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test3",
				Namespace: "thin",
			},
			Spec: datav1alpha1.ThinRuntimeSpec{
				Worker: datav1alpha1.ThinCompTemplateSpec{
					Enabled: true,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				WorkerPhase: datav1alpha1.RuntimePhaseReady,
			},
		}

		data := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test3",
				Namespace: "thin",
			},
		}

		s := apimachineryruntime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, runtime, data)
		_ = v1.AddToScheme(s)

		mockClient := fake.NewFakeClientWithScheme(s, runtime, data)

		e := &ThinEngine{
			name:      "test3",
			namespace: "thin",
			runtime:   runtime,
			Client:    mockClient,
		}

		gotShould, err := e.ShouldSetupWorkers()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotShould).To(BeFalse())
	})

	It("should return false when worker is not enabled", func() {
		runtime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test3",
				Namespace: "thin",
			},
		}

		data := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test3",
				Namespace: "thin",
			},
		}

		s := apimachineryruntime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, runtime, data)
		_ = v1.AddToScheme(s)

		mockClient := fake.NewFakeClientWithScheme(s, runtime, data)

		e := &ThinEngine{
			name:      "test3",
			namespace: "thin",
			runtime:   runtime,
			Client:    mockClient,
		}

		gotShould, err := e.ShouldSetupWorkers()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotShould).To(BeFalse())
	})
})

var _ = Describe("ThinEngine_SetupWorkers", func() {
	It("should setup workers and scale correctly", func() {
		runtimeInfo, err := base.BuildRuntimeInfo("thin", "fluid", common.ThinRuntime)
		Expect(err).NotTo(HaveOccurred())

		runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
		})

		nodeSelector := map[string]string{
			"node-select": "true",
		}
		runtimeInfo.SetFuseNodeSelector(nodeSelector)

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		}

		worker := appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
		}

		runtime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.ThinRuntimeSpec{
				Replicas: 1,
				Worker: datav1alpha1.ThinCompTemplateSpec{
					Enabled: true,
				},
			},
		}

		s := apimachineryruntime.NewScheme()
		data := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
		}
		s.AddKnownTypes(datav1alpha1.GroupVersion, runtime, data)
		s.AddKnownTypes(appsv1.SchemeGroupVersion, &worker)
		_ = v1.AddToScheme(s)

		mockClient := fake.NewFakeClientWithScheme(s, node.DeepCopy(), worker.DeepCopy(), runtime, data)

		e := &ThinEngine{
			runtime:     runtime,
			runtimeInfo: runtimeInfo,
			Client:      mockClient,
			name:        "test",
			namespace:   "fluid",
			Log:         ctrl.Log.WithName("test"),
		}

		e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

		err = e.SetupWorkers()
		Expect(err).NotTo(HaveOccurred())

		var updatedWorker appsv1.StatefulSet
		err = mockClient.Get(context.TODO(), client.ObjectKey{Name: "test-worker", Namespace: "fluid"}, &updatedWorker)
		Expect(err).NotTo(HaveOccurred())
		Expect(*updatedWorker.Spec.Replicas).To(Equal(int32(1)))
	})
})

var _ = Describe("ThinEngine_CheckWorkersReady", func() {
	It("should return true when workers and fuse are ready", func() {
		runtime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test0",
				Namespace: "thin",
			},
			Spec: datav1alpha1.ThinRuntimeSpec{
				Replicas: 1,
				Worker: datav1alpha1.ThinCompTemplateSpec{
					Enabled: true,
				},
				Fuse: datav1alpha1.ThinFuseSpec{},
			},
		}

		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test0-worker",
				Namespace: "thin",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		}

		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test0-fuse",
				Namespace: "thin",
			},
			Status: appsv1.DaemonSetStatus{
				NumberAvailable:        1,
				DesiredNumberScheduled: 1,
				CurrentNumberScheduled: 1,
			},
		}

		data := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test0",
				Namespace: "thin",
			},
		}

		s := apimachineryruntime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, runtime, data)
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse, worker)
		_ = v1.AddToScheme(s)

		mockClient := fake.NewFakeClientWithScheme(s, runtime, data, fuse, worker)

		e := &ThinEngine{
			runtime:   runtime,
			name:      "test0",
			namespace: "thin",
			Client:    mockClient,
			Log:       ctrl.Log.WithName("test0"),
		}

		runtimeInfo, err := base.BuildRuntimeInfo("test0", "thin", common.ThinRuntime)
		Expect(err).NotTo(HaveOccurred())

		e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

		gotReady, err := e.CheckWorkersReady()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotReady).To(BeTrue())
	})

	It("should return false when workers are not ready", func() {
		runtime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "thin",
			},
			Spec: datav1alpha1.ThinRuntimeSpec{
				Replicas: 1,
				Worker: datav1alpha1.ThinCompTemplateSpec{
					Enabled: true,
				},
				Fuse: datav1alpha1.ThinFuseSpec{},
			},
		}

		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1-worker",
				Namespace: "thin",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 0,
			},
		}

		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1-fuse",
				Namespace: "thin",
			},
			Status: appsv1.DaemonSetStatus{
				NumberAvailable:        0,
				DesiredNumberScheduled: 1,
				CurrentNumberScheduled: 0,
			},
		}

		data := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "thin",
			},
		}

		s := apimachineryruntime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, runtime, data)
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse, worker)
		_ = v1.AddToScheme(s)

		mockClient := fake.NewFakeClientWithScheme(s, runtime, data, fuse, worker)

		e := &ThinEngine{
			runtime:   runtime,
			name:      "test1",
			namespace: "thin",
			Client:    mockClient,
			Log:       ctrl.Log.WithName("test1"),
		}

		runtimeInfo, err := base.BuildRuntimeInfo("test1", "thin", common.ThinRuntime)
		Expect(err).NotTo(HaveOccurred())

		e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

		gotReady, err := e.CheckWorkersReady()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotReady).To(BeFalse())
	})
})

var _ = Describe("ThinEngine_GetWorkerSelectors", func() {
	It("should return correct worker selector", func() {
		e := &ThinEngine{
			name: "spark",
		}

		got := e.getWorkerSelectors()

		Expect(got).To(Equal("app=thin,release=spark,role=thin-worker"))
	})
})
