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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	healthCheckTestNamespace       = "fluid"
	healthCheckTestNameHbase       = "hbase"
	healthCheckTestNameTest        = "test"
	healthCheckTestNameSpark       = "spark"
	healthCheckTestWorkerSuffix    = "-worker"
	healthCheckTestFuseSuffix      = "-fuse"
	healthCheckTestEndpoint        = "test Endpoint"
	healthCheckTestHCFSVersion     = "Underlayer HCFS Compatible Version"
	healthCheckTestCachedValue     = "true"
)

var _ = Describe("CheckRuntimeHealthy", Label("pkg.ddc.juicefs.health_check_test.go"), func() {
	var (
		client        client.Client
		testObjs      []runtime.Object
		stsInputs     []appsv1.StatefulSet
		dsInputs      []appsv1.DaemonSet
		runtimeInputs []datav1alpha1.JuiceFSRuntime
		datasetInputs []*datav1alpha1.Dataset
	)

	BeforeEach(func() {
		stsInputs = []appsv1.StatefulSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameHbase + healthCheckTestWorkerSuffix,
					Namespace: healthCheckTestNamespace,
				},
				Status: appsv1.StatefulSetStatus{
					Replicas:          1,
					ReadyReplicas:     1,
					AvailableReplicas: 1,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameTest + healthCheckTestWorkerSuffix,
					Namespace: healthCheckTestNamespace,
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

		dsInputs = []appsv1.DaemonSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameHbase + healthCheckTestFuseSuffix,
					Namespace: healthCheckTestNamespace,
				},
				Status: appsv1.DaemonSetStatus{
					NumberUnavailable: 0,
					NumberReady:       1,
					NumberAvailable:   1,
				},
			},
		}

		runtimeInputs = []datav1alpha1.JuiceFSRuntime{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameHbase,
					Namespace: healthCheckTestNamespace,
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Replicas: 1,
				},
				Status: datav1alpha1.RuntimeStatus{
					CacheStates: map[common.CacheStateName]string{
						common.Cached: healthCheckTestCachedValue,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameTest,
					Namespace: healthCheckTestNamespace,
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Replicas: 1,
				},
				Status: datav1alpha1.RuntimeStatus{
					CacheStates: map[common.CacheStateName]string{
						common.Cached: healthCheckTestCachedValue,
					},
				},
			},
		}

		datasetInputs = []*datav1alpha1.Dataset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameHbase,
					Namespace: healthCheckTestNamespace,
				},
				Spec: datav1alpha1.DatasetSpec{},
				Status: datav1alpha1.DatasetStatus{
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    healthCheckTestEndpoint,
						UnderlayerFileSystemVersion: healthCheckTestHCFSVersion,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameTest,
					Namespace: healthCheckTestNamespace,
				},
				Spec: datav1alpha1.DatasetSpec{},
				Status: datav1alpha1.DatasetStatus{
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    healthCheckTestEndpoint,
						UnderlayerFileSystemVersion: healthCheckTestHCFSVersion,
					},
				},
			},
		}

		testObjs = []runtime.Object{}
		for _, ds := range dsInputs {
			testObjs = append(testObjs, ds.DeepCopy())
		}
		for _, sts := range stsInputs {
			testObjs = append(testObjs, sts.DeepCopy())
		}
		for _, rt := range runtimeInputs {
			testObjs = append(testObjs, rt.DeepCopy())
		}
		for _, ds := range datasetInputs {
			testObjs = append(testObjs, ds.DeepCopy())
		}

		client = fake.NewFakeClientWithScheme(testScheme, testObjs...)
	})

	Context("when runtime is healthy", func() {
		It("should update runtime status correctly", func() {
			engine := JuiceFSEngine{
				Client:    client,
				Log:       fake.NullLogger(),
				namespace: healthCheckTestNamespace,
				name:      healthCheckTestNameHbase,
				runtime:   &runtimeInputs[0],
				Recorder:  record.NewFakeRecorder(1),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
			Expect(err).NotTo(HaveOccurred())
			engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)

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
			err = client.List(context.TODO(), &datasets)
			Expect(err).NotTo(HaveOccurred())
			Expect(datasets.Items[0].Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			Expect(datasets.Items[0].Status.CacheStates).To(HaveKeyWithValue(common.Cached, healthCheckTestCachedValue))
			Expect(datasets.Items[0].Status.HCFSStatus.Endpoint).To(Equal(healthCheckTestEndpoint))
			Expect(datasets.Items[0].Status.HCFSStatus.UnderlayerFileSystemVersion).To(Equal(healthCheckTestHCFSVersion))
		})
	})

	Context("when worker is not ready", func() {
		It("should return error", func() {
			engine := JuiceFSEngine{
				Client:    client,
				Log:       fake.NullLogger(),
				namespace: healthCheckTestNamespace,
				name:      healthCheckTestNameTest,
				runtime:   &runtimeInputs[1],
				Recorder:  record.NewFakeRecorder(1),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
			Expect(err).NotTo(HaveOccurred())
			engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)

			err = engine.CheckRuntimeHealthy()
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("CheckFuseHealthy", Label("pkg.ddc.juicefs.health_check_test.go"), func() {
	var (
		client        client.Client
		testObjs      []runtime.Object
		dsInputs      []appsv1.DaemonSet
		runtimeInputs []datav1alpha1.JuiceFSRuntime
	)

	BeforeEach(func() {
		dsInputs = []appsv1.DaemonSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameHbase + healthCheckTestFuseSuffix,
					Namespace: healthCheckTestNamespace,
				},
				Status: appsv1.DaemonSetStatus{
					NumberUnavailable: 1,
					NumberReady:       1,
					NumberAvailable:   1,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameSpark + healthCheckTestFuseSuffix,
					Namespace: healthCheckTestNamespace,
				},
				Status: appsv1.DaemonSetStatus{
					NumberUnavailable: 0,
					NumberReady:       1,
					NumberAvailable:   1,
				},
			},
		}

		runtimeInputs = []datav1alpha1.JuiceFSRuntime{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameHbase,
					Namespace: healthCheckTestNamespace,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestNameSpark,
					Namespace: healthCheckTestNamespace,
				},
			},
		}

		testObjs = []runtime.Object{}
		for _, ds := range dsInputs {
			testObjs = append(testObjs, ds.DeepCopy())
		}
		for _, rt := range runtimeInputs {
			testObjs = append(testObjs, rt.DeepCopy())
		}

		client = fake.NewFakeClientWithScheme(testScheme, testObjs...)
	})

	Context("when fuse has unavailable pods", func() {
		It("should set runtime to NotReady phase", func() {
			engine := JuiceFSEngine{
				Client:    client,
				Log:       fake.NullLogger(),
				namespace: healthCheckTestNamespace,
				name:      healthCheckTestNameHbase,
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestNameHbase,
						Namespace: healthCheckTestNamespace,
					},
				},
				Recorder: record.NewFakeRecorder(1),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
			Expect(err).NotTo(HaveOccurred())
			engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)

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
	})

	Context("when fuse is fully ready", func() {
		It("should set runtime to Ready phase", func() {
			engine := JuiceFSEngine{
				Client:    client,
				Log:       fake.NullLogger(),
				namespace: healthCheckTestNamespace,
				name:      healthCheckTestNameSpark,
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestNameSpark,
						Namespace: healthCheckTestNamespace,
					},
				},
				Recorder: record.NewFakeRecorder(1),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
			Expect(err).NotTo(HaveOccurred())
			engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)

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
})
