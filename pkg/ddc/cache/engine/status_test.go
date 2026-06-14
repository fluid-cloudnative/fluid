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

package engine

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	testStatusNamespace  = "default"
	testStatusRuntime    = "curvine-demo"
	testStatusMaster     = "curvine-demo-master"
	testStatusWorker     = "curvine-demo-worker"
	testStatusClient     = "curvine-demo-client"
	testCacheRuntimeGR   = "cacheruntimes"
	testCacheRuntimeGV   = "data.fluid.io"
	testStatusWorkloadAP = "apps/v1"
)

var _ = Describe("CheckAndUpdateRuntimeStatus", func() {
	var (
		engine *CacheEngine
		client ctrlclient.Client
	)

	BeforeEach(func() {
		// Default setup can be added here if needed
	})

	Describe("Client component status handling", func() {
		Context("when client is not ready", func() {
			BeforeEach(func() {
				engine, client = newStatusTestEngineWithClient(
					fake.NewFakeClientWithScheme(
						CacheEngineTestScheme,
						newStatusTestRuntime(),
						newAdvancedStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
						newAdvancedStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
						newDaemonSetComponent(testStatusClient, testStatusNamespace, 1, 0),
					),
				)
			})

			It("should not block runtime ready and set client phase to NotReady", func() {
				ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(true))
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue(), "expected runtime to become ready once master and worker are ready")

				updatedRuntime := getUpdatedRuntime(client)
				Expect(updatedRuntime.Status.Client.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedRuntime.Status.SetupDuration).NotTo(BeEmpty(), "expected setup duration to be recorded once runtime is ready")
			})
		})

		Context("when client is partially ready", func() {
			BeforeEach(func() {
				engine, client = newStatusTestEngineWithClient(
					fake.NewFakeClientWithScheme(
						CacheEngineTestScheme,
						newStatusTestRuntime(),
						newAdvancedStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
						newAdvancedStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
						newDaemonSetComponent(testStatusClient, testStatusNamespace, 2, 1),
					),
				)
			})

			It("should not block runtime ready and set client phase to PartialReady", func() {
				ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(true))
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue(), "expected runtime to become ready once master and worker are ready")

				updatedRuntime := getUpdatedRuntime(client)
				Expect(updatedRuntime.Status.Client.Phase).To(Equal(datav1alpha1.RuntimePhasePartialReady))
				Expect(updatedRuntime.Status.SetupDuration).NotTo(BeEmpty(), "expected setup duration to be recorded once runtime is ready")
			})
		})

		Context("when client has zero desired replicas", func() {
			BeforeEach(func() {
				engine, client = newStatusTestEngineWithClient(
					fake.NewFakeClientWithScheme(
						CacheEngineTestScheme,
						newStatusTestRuntime(),
						newAdvancedStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
						newAdvancedStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
						newDaemonSetComponent(testStatusClient, testStatusNamespace, 0, 0),
					),
				)
			})

			It("should keep runtime ready and set client phase to NotReady with zero desired replicas", func() {
				ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(true))
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue(), "expected runtime to stay ready when client desires zero replicas")

				updatedRuntime := getUpdatedRuntime(client)
				Expect(updatedRuntime.Status.Client.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedRuntime.Status.Client.DesiredReplicas).To(Equal(int32(0)), "expected desired replicas to stay 0")
			})
		})

		Context("when client is fully ready", func() {
			BeforeEach(func() {
				engine, client = newStatusTestEngineWithClient(
					fake.NewFakeClientWithScheme(
						CacheEngineTestScheme,
						newStatusTestRuntime(),
						newAdvancedStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
						newAdvancedStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
						newDaemonSetComponent(testStatusClient, testStatusNamespace, 2, 2),
					),
				)
			})

			It("should keep runtime ready and set client phase to Ready", func() {
				ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(true))
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue(), "expected runtime to stay ready when client is fully ready")

				updatedRuntime := getUpdatedRuntime(client)
				Expect(updatedRuntime.Status.Client.Phase).To(Equal(datav1alpha1.RuntimePhaseReady))
				Expect(updatedRuntime.Status.Client.ReadyReplicas).To(Equal(updatedRuntime.Status.Client.DesiredReplicas),
					"expected ready replicas to match desired replicas")
			})
		})
	})

	Describe("Runtime status recomputation on retry", func() {
		Context("when conflict occurs during status update", func() {
			It("should recompute runtime ready status after retry", func() {
				baseClient := fake.NewFakeClientWithScheme(
					CacheEngineTestScheme,
					newStatusTestRuntime(),
					newAdvancedStatefulSetComponent(testStatusMaster, testStatusNamespace, 1, 1),
					newAdvancedStatefulSetComponent(testStatusWorker, testStatusNamespace, 1, 1),
				)

				client := &conflictOnceClient{
					Client: baseClient,
					statusWriter: &conflictOnceStatusWriter{
						StatusWriter: baseClient.Status(),
						beforeConflict: func(ctx context.Context) error {
							worker := &workloadv1alpha1.AdvancedStatefulSet{}
							if err := baseClient.Get(ctx, types.NamespacedName{Name: testStatusWorker, Namespace: testStatusNamespace}, worker); err != nil {
								return err
							}

							worker.Status.ReadyReplicas = 0
							worker.Status.AvailableReplicas = 0
							return baseClient.Status().Update(ctx, worker)
						},
					},
				}

				engine, _ := newStatusTestEngineWithClient(client)
				ready, err := engine.CheckAndUpdateRuntimeStatus(newStatusTestRuntimeValue(false))
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeFalse(), "expected runtime to be not ready after retry sees worker become not ready")

				updatedRuntime := getUpdatedRuntime(client)
				Expect(updatedRuntime.Status.Worker.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedRuntime.Status.SetupDuration).To(BeEmpty(),
					"expected setup duration to stay empty when final runtime status is not ready")
			})
		})
	})
})

func newStatusTestEngineWithClient(client ctrlclient.Client) (*CacheEngine, ctrlclient.Client) {
	return &CacheEngine{
		Client:    client,
		name:      testStatusRuntime,
		namespace: testStatusNamespace,
		Log:       fake.NullLogger(),
	}, client
}

func newStatusTestRuntimeValue(enableClient bool) *common.CacheRuntimeStatusValue {
	value := &common.CacheRuntimeStatusValue{
		Master: newStatusTestComponentStatusInfo(common.ComponentTypeMaster, testStatusMaster),
		Worker: newStatusTestComponentStatusInfo(common.ComponentTypeWorker, testStatusWorker),
		Client: newStatusTestComponentStatusInfo(common.ComponentTypeClient, testStatusClient),
	}
	value.Client.Enabled = enableClient

	return value
}

func newStatusTestComponentStatusInfo(componentType common.ComponentType, name string) *common.ComponentStatusInfo {
	return &common.ComponentStatusInfo{
		ComponentIdentity: common.ComponentIdentity{
			Name:      name,
			Namespace: testStatusNamespace,
		},
		Enabled: true,
	}
}

func newStatusTestRuntime() *datav1alpha1.CacheRuntime {
	return &datav1alpha1.CacheRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:              testStatusRuntime,
			Namespace:         testStatusNamespace,
			CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Minute)),
		},
	}
}

func getUpdatedRuntime(client ctrlclient.Client) *datav1alpha1.CacheRuntime {
	updatedRuntime := &datav1alpha1.CacheRuntime{}
	Expect(client.Get(context.TODO(), types.NamespacedName{Name: testStatusRuntime, Namespace: testStatusNamespace}, updatedRuntime)).To(Succeed())
	return updatedRuntime
}

func newAdvancedStatefulSetComponent(name, namespace string, desiredReplicas, readyReplicas int32) *workloadv1alpha1.AdvancedStatefulSet {
	replicas := desiredReplicas
	return &workloadv1alpha1.AdvancedStatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
			Replicas: &replicas,
		},
		Status: workloadv1alpha1.AdvancedStatefulSetStatus{
			CurrentReplicas:   desiredReplicas,
			AvailableReplicas: readyReplicas,
			ReadyReplicas:     readyReplicas,
		},
	}
}

func newDaemonSetComponent(name, namespace string, desiredReplicas, readyReplicas int32) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: appsv1.DaemonSetStatus{
			CurrentNumberScheduled: desiredReplicas,
			DesiredNumberScheduled: desiredReplicas,
			NumberAvailable:        readyReplicas,
			NumberReady:            readyReplicas,
			NumberUnavailable:      desiredReplicas - readyReplicas,
		},
	}
}

type conflictOnceClient struct {
	ctrlclient.Client
	statusWriter ctrlclient.StatusWriter
}

func (c *conflictOnceClient) Status() ctrlclient.StatusWriter {
	return c.statusWriter
}

type conflictOnceStatusWriter struct {
	ctrlclient.StatusWriter
	beforeConflict func(ctx context.Context) error
	conflicted     bool
}

func (w *conflictOnceStatusWriter) Update(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.SubResourceUpdateOption) error {
	if !w.conflicted {
		w.conflicted = true
		if w.beforeConflict != nil {
			if err := w.beforeConflict(ctx); err != nil {
				return err
			}
		}

		return apierrors.NewConflict(
			schema.GroupResource{Group: testCacheRuntimeGV, Resource: testCacheRuntimeGR},
			obj.GetName(),
			errors.New("injected conflict"),
		)
	}

	return w.StatusWriter.Update(ctx, obj, opts...)
}
