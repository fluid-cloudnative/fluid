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
package juicefs

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/utils/ptr"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("CheckRuntimeHealthy", func() {
	var (
		stsInputs            []appsv1.StatefulSet
		daemonSetInputs      []appsv1.DaemonSet
		juicefsruntimeInputs []datav1alpha1.JuiceFSRuntime
		datasetInputs        []*datav1alpha1.Dataset
		testObjs             []runtime.Object
		fakeClient           client.Client
		engines              []JuiceFSEngine
	)

	BeforeEach(func() {
		stsInputs = []appsv1.StatefulSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-worker",
					Namespace: "fluid",
				},
				Status: appsv1.StatefulSetStatus{
					Replicas:          1,
					ReadyReplicas:     1,
					AvailableReplicas: 1,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-worker",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
				},
				Status: appsv1.StatefulSetStatus{
					Replicas:          1,
					ReadyReplicas:     0,
					AvailableReplicas: 0,
				},
			},
		}

		daemonSetInputs = []appsv1.DaemonSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-fuse",
					Namespace: "fluid",
				},
				Status: appsv1.DaemonSetStatus{
					NumberUnavailable: 0,
					NumberReady:       1,
					NumberAvailable:   1,
				},
			},
		}

		juicefsruntimeInputs = []datav1alpha1.JuiceFSRuntime{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Replicas: 1,
				},
				Status: datav1alpha1.RuntimeStatus{
					CacheStates: map[common.CacheStateName]string{
						common.Cached: "true",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Replicas: 1,
				},
				Status: datav1alpha1.RuntimeStatus{
					CacheStates: map[common.CacheStateName]string{
						common.Cached: "true",
					},
				},
			},
		}

		datasetInputs = []*datav1alpha1.Dataset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{},
				Status: datav1alpha1.DatasetStatus{
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    "test Endpoint",
						UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{},
				Status: datav1alpha1.DatasetStatus{
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    "test Endpoint",
						UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
					},
				},
			},
		}

		testObjs = []runtime.Object{}
		for _, daemonSet := range daemonSetInputs {
			testObjs = append(testObjs, daemonSet.DeepCopy())
		}
		for _, sts := range stsInputs {
			testObjs = append(testObjs, sts.DeepCopy())
		}
		for _, juicefsruntime := range juicefsruntimeInputs {
			testObjs = append(testObjs, juicefsruntime.DeepCopy())
		}
		for _, dataset := range datasetInputs {
			testObjs = append(testObjs, dataset.DeepCopy())
		}

		fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

		engines = []JuiceFSEngine{
			{
				Client:    fakeClient,
				Log:       fake.NullLogger(),
				namespace: "fluid",
				name:      "hbase",
				runtime:   &juicefsruntimeInputs[0],
				Recorder:  record.NewFakeRecorder(1),
			},
			{
				Client:    fakeClient,
				Log:       fake.NullLogger(),
				namespace: "fluid",
				name:      "test",
				runtime:   &juicefsruntimeInputs[1],
				Recorder:  record.NewFakeRecorder(1),
			},
		}
	})

	It("should succeed with healthy runtime", func() {
		engine := engines[0]
		runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
		Expect(err).NotTo(HaveOccurred())
		engine.Helper = ctrl.BuildHelper(runtimeInfo, fakeClient, engine.Log)

		err = engine.CheckRuntimeHealthy()
		Expect(err).NotTo(HaveOccurred())

		juicefsruntime, err := engine.getRuntime()
		Expect(err).NotTo(HaveOccurred())
		Expect(juicefsruntime.Status.WorkerNumberReady).To(Equal(int32(1)))
		Expect(juicefsruntime.Status.WorkerNumberAvailable).To(Equal(int32(1)))
		Expect(juicefsruntime.Status.FuseNumberReady).To(Equal(int32(1)))
		Expect(juicefsruntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
		Expect(juicefsruntime.Status.FuseNumberUnavailable).To(Equal(int32(0)))

		_, cond := utils.GetRuntimeCondition(juicefsruntime.Status.Conditions, datav1alpha1.RuntimeWorkersReady)
		Expect(cond).NotTo(BeNil())
		_, cond = utils.GetRuntimeCondition(juicefsruntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
		Expect(cond).NotTo(BeNil())

		var datasets datav1alpha1.DatasetList
		err = fakeClient.List(context.TODO(), &datasets)
		Expect(err).NotTo(HaveOccurred())
		Expect(datasets.Items[0].Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
		Expect(datasets.Items[0].Status.CacheStates).To(BeEquivalentTo(map[common.CacheStateName]string{
			common.Cached: "true",
		}))
		Expect(datasets.Items[0].Status.HCFSStatus).To(Equal(&datav1alpha1.HCFSStatus{
			Endpoint:                    "test Endpoint",
			UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
		}))
	})

	It("should fail with unhealthy runtime", func() {
		engine := engines[1]
		runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
		Expect(err).NotTo(HaveOccurred())
		engine.Helper = ctrl.BuildHelper(runtimeInfo, fakeClient, engine.Log)

		err = engine.CheckRuntimeHealthy()
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("CheckFuseHealthy", func() {
	var (
		daemonSetInputs      []appsv1.DaemonSet
		juicefsruntimeInputs []datav1alpha1.JuiceFSRuntime
		testObjs             []runtime.Object
		fakeClient           client.Client
		engines              []JuiceFSEngine
	)

	BeforeEach(func() {
		daemonSetInputs = []appsv1.DaemonSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-fuse",
					Namespace: "fluid",
				},
				Status: appsv1.DaemonSetStatus{
					NumberUnavailable: 1,
					NumberReady:       1,
					NumberAvailable:   1,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark-fuse",
					Namespace: "fluid",
				},
				Status: appsv1.DaemonSetStatus{
					NumberUnavailable: 0,
					NumberReady:       1,
					NumberAvailable:   1,
				},
			},
		}

		juicefsruntimeInputs = []datav1alpha1.JuiceFSRuntime{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
			},
		}

		testObjs = []runtime.Object{}
		for _, daemonSet := range daemonSetInputs {
			testObjs = append(testObjs, daemonSet.DeepCopy())
		}
		for _, juicefsruntimeInput := range juicefsruntimeInputs {
			testObjs = append(testObjs, juicefsruntimeInput.DeepCopy())
		}

		fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

		engines = []JuiceFSEngine{
			{
				Client:    fakeClient,
				Log:       fake.NullLogger(),
				namespace: "fluid",
				name:      "hbase",
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
				},
				Recorder: record.NewFakeRecorder(1),
			},
			{
				Client:    fakeClient,
				Log:       fake.NullLogger(),
				namespace: "fluid",
				name:      "spark",
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "fluid",
					},
				},
				Recorder: record.NewFakeRecorder(1),
			},
		}
	})

	It("should update status with unavailable fuses", func() {
		engine := engines[0]
		runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
		Expect(err).NotTo(HaveOccurred())
		engine.Helper = ctrl.BuildHelper(runtimeInfo, fakeClient, engine.Log)

		_, err = engine.checkFuseHealthy()
		Expect(err).NotTo(HaveOccurred())

		juicefsruntime, err := engine.getRuntime()
		Expect(err).NotTo(HaveOccurred())
		Expect(juicefsruntime.Status.FuseNumberReady).To(Equal(int32(1)))
		Expect(juicefsruntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
		Expect(juicefsruntime.Status.FuseNumberUnavailable).To(Equal(int32(1)))

		_, cond := utils.GetRuntimeCondition(juicefsruntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
		Expect(cond).NotTo(BeNil())
	})

	It("should update status with all fuses available", func() {
		engine := engines[1]
		runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
		Expect(err).NotTo(HaveOccurred())
		engine.Helper = ctrl.BuildHelper(runtimeInfo, fakeClient, engine.Log)

		_, err = engine.checkFuseHealthy()
		Expect(err).NotTo(HaveOccurred())

		juicefsruntime, err := engine.getRuntime()
		Expect(err).NotTo(HaveOccurred())
		Expect(juicefsruntime.Status.FuseNumberReady).To(Equal(int32(1)))
		Expect(juicefsruntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
		Expect(juicefsruntime.Status.FuseNumberUnavailable).To(Equal(int32(0)))

		_, cond := utils.GetRuntimeCondition(juicefsruntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
		Expect(cond).NotTo(BeNil())
	})
})
