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

package alluxio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/features"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
)

const testDrainWorkerSts = "drain-worker"
const testDrainNamespace = "fluid"

var _ = Describe("AlluxioEngine drainScalingDownWorkers", Label("pkg.ddc.alluxio.replicas_drain_test.go"), func() {
	var (
		engine *AlluxioEngine
		rt     *v1alpha1.AlluxioRuntime
	)

	BeforeEach(func() {
		rt = &v1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testDrainWorkerSts,
				Namespace: testDrainNamespace,
			},
		}
	})

	newEngineWithPods := func(pods ...*corev1.Pod) *AlluxioEngine {
		objs := []runtime.Object{}
		for _, p := range pods {
			objs = append(objs, p.DeepCopy())
		}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
		return newAlluxioEngineREP(fakeClient, testDrainWorkerSts, testDrainNamespace)
	}

	// hostIP mirrors status.hostIP, which is what ALLUXIO_WORKER_HOSTNAME (and
	// therefore the worker's registered identity with the master) is sourced
	// from in charts/alluxio - not the pod's own IP.
	workerPod := func(ordinal int, hostIP string) *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-worker-%d", testDrainWorkerSts, ordinal),
				Namespace: testDrainNamespace,
			},
			Status: corev1.PodStatus{
				HostIP: hostIP,
			},
		}
	}

	Context("when the pod targeted for removal is already gone", func() {
		It("treats a NotFound pod as already decommissioned", func() {
			engine = newEngineWithPods()
			drained, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
		})
	})

	Context("when the pod has not yet been assigned a host IP", func() {
		It("returns not drained without error", func() {
			engine = newEngineWithPods(workerPod(1, ""))
			drained, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeFalse())
		})
	})

	Context("when the decommission call fails", func() {
		It("propagates the error", func() {
			engine = newEngineWithPods(workerPod(1, "10.0.0.1"))
			patch := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error {
					return errors.New("decommission failed")
				})
			defer patch.Reset()

			drained, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).To(HaveOccurred())
			Expect(drained).To(BeFalse())
		})
	})

	Context("when active workers are still above the desired count", func() {
		It("returns not drained and requests a retry", func() {
			engine = newEngineWithPods(workerPod(1, "10.0.0.1"))
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error {
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) {
					return 2, nil
				})
			defer patch2.Reset()

			drained, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeFalse())
		})
	})

	Context("when the worker has successfully drained", func() {
		It("returns drained with no error", func() {
			engine = newEngineWithPods(workerPod(1, "10.0.0.1"))
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error {
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) {
					return 1, nil
				})
			defer patch2.Reset()

			drained, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
		})
	})

	Context("when multiple targeted pods share the same node", func() {
		It("deduplicates the decommission address list", func() {
			engine = newEngineWithPods(workerPod(1, "10.0.0.1"), workerPod(2, "10.0.0.1"))

			var capturedAddrs []string
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, addrs []string) error {
					capturedAddrs = append([]string(nil), addrs...)
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) {
					return 1, nil
				})
			defer patch2.Reset()

			drained, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
			Expect(capturedAddrs).To(HaveLen(1))
		})
	})
})

var _ = Describe("AlluxioEngine getWorkerRPCPort", Label("pkg.ddc.alluxio.replicas_drain_test.go"), func() {
	var engine *AlluxioEngine

	BeforeEach(func() {
		engine = newAlluxioEngineREP(fake.NewFakeClientWithScheme(testScheme), testDrainWorkerSts, testDrainNamespace)
	})

	It("returns the configured rpc port when set", func() {
		rt := &v1alpha1.AlluxioRuntime{
			Spec: v1alpha1.AlluxioRuntimeSpec{
				Worker: v1alpha1.AlluxioCompTemplateSpec{
					Ports: map[string]int{"rpc": 12345},
				},
			},
		}
		Expect(engine.getWorkerRPCPort(rt)).To(Equal(12345))
	})

	It("falls back to the default port when unset", func() {
		rt := &v1alpha1.AlluxioRuntime{}
		Expect(engine.getWorkerRPCPort(rt)).To(Equal(defaultWorkerRPCPort))
	})

	It("falls back to the default port when the configured value is not positive", func() {
		rt := &v1alpha1.AlluxioRuntime{
			Spec: v1alpha1.AlluxioRuntimeSpec{
				Worker: v1alpha1.AlluxioCompTemplateSpec{
					Ports: map[string]int{"rpc": 0},
				},
			},
		}
		Expect(engine.getWorkerRPCPort(rt)).To(Equal(defaultWorkerRPCPort))
	})
})

var _ = Describe("AlluxioEngine SyncReplicas worker decommission deadline", Label("pkg.ddc.alluxio.replicas_drain_test.go"), func() {
	const (
		deadlineTestRuntime = "deadline-worker"
		deadlineTestNs      = "fluid"
	)

	newFixtures := func(existingCond *v1alpha1.RuntimeCondition) *AlluxioEngine {
		rt := &v1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime, Namespace: deadlineTestNs},
			Spec:       v1alpha1.AlluxioRuntimeSpec{Replicas: 1},
			Status:     v1alpha1.RuntimeStatus{DesiredWorkerNumberScheduled: 2},
		}
		if existingCond != nil {
			rt.Status.Conditions = []v1alpha1.RuntimeCondition{*existingCond}
		}
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime + "-worker", Namespace: deadlineTestNs},
			Spec:       appsv1.StatefulSetSpec{Replicas: ptr.To[int32](2)},
			Status:     appsv1.StatefulSetStatus{Replicas: 2},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime + "-worker-1", Namespace: deadlineTestNs},
			Status:     corev1.PodStatus{HostIP: "10.0.0.5"},
		}
		// BuildWorkersAffinity (invoked when Helper.SyncReplicas updates the
		// StatefulSet's replica count) requires the Dataset to exist.
		dataset := &v1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime, Namespace: deadlineTestNs},
		}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, rt, sts, pod, dataset)
		return newAlluxioEngineREP(fakeClient, deadlineTestRuntime, deadlineTestNs)
	}

	getCondition := func(engine *AlluxioEngine) *v1alpha1.RuntimeCondition {
		rt, err := engine.getRuntime()
		Expect(err).NotTo(HaveOccurred())
		_, cond := utils.GetRuntimeCondition(rt.Status.Conditions, v1alpha1.RuntimeWorkerDecommissioning)
		return cond
	}

	BeforeEach(func() {
		Expect(utilfeature.DefaultMutableFeatureGate.Set(string(features.GracefulWorkerScaleDown) + "=true")).To(Succeed())
	})

	AfterEach(func() {
		Expect(utilfeature.DefaultMutableFeatureGate.Set(string(features.GracefulWorkerScaleDown) + "=false")).To(Succeed())
	})

	Context("when a drain doesn't finish within one reconcile", func() {
		It("records when the decommission attempt started and keeps requeuing", func() {
			engine := newFixtures(nil)
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error { return nil })
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) { return 2, nil })
			defer patch2.Reset()

			err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
				Log: fake.NullLogger(), Recorder: record.NewFakeRecorder(300),
			})
			Expect(errors.Is(err, errWorkersNotYetDrained)).To(BeTrue())

			cond := getCondition(engine)
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			Expect(time.Since(cond.LastTransitionTime.Time)).To(BeNumerically("<", time.Minute))
		})
	})

	Context("when a drain is still stuck past the deadline", func() {
		It("forces the scale-down to proceed and clears the marker", func() {
			staleCond := utils.NewRuntimeCondition(v1alpha1.RuntimeWorkerDecommissioning,
				v1alpha1.RuntimeWorkerDecommissioningReason, "started earlier", corev1.ConditionTrue)
			staleCond.LastTransitionTime = metav1.NewTime(time.Now().Add(-defaultWorkerDecommissionDeadline - time.Minute))
			engine := newFixtures(&staleCond)

			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error { return nil })
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) { return 2, nil })
			defer patch2.Reset()

			err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
				Log: fake.NullLogger(), Recorder: record.NewFakeRecorder(300),
			})
			Expect(err).NotTo(HaveOccurred())

			cond := getCondition(engine)
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(corev1.ConditionFalse))

			var sts appsv1.StatefulSet
			Expect(engine.Client.Get(context.TODO(),
				types.NamespacedName{Name: deadlineTestRuntime + "-worker", Namespace: deadlineTestNs}, &sts)).To(Succeed())
			Expect(*sts.Spec.Replicas).To(Equal(int32(1)))
		})
	})
})
