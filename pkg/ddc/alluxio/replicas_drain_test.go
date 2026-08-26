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
		phase := corev1.PodRunning
		if hostIP == "" {
			phase = corev1.PodPending
		}
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-worker-%d", testDrainWorkerSts, ordinal),
				Namespace: testDrainNamespace,
			},
			Status: corev1.PodStatus{
				HostIP: hostIP,
				Phase:  phase,
			},
		}
	}

	// workerPodWithWebPort mirrors what charts/alluxio/templates/worker/
	// statefulset.yaml actually renders onto the alluxio-worker container -
	// used to prove the decommission address is read from the pod's own
	// bound port rather than assumed from the runtime spec.
	workerPodWithWebPort := func(ordinal int, hostIP string, webPort int32) *corev1.Pod {
		pod := workerPod(ordinal, hostIP)
		pod.Spec.Containers = []corev1.Container{
			{
				Name: alluxioWorkerContainerName,
				Ports: []corev1.ContainerPort{
					{Name: "web", ContainerPort: webPort},
				},
			},
		}
		return pod
	}

	terminatingWorkerPod := func(ordinal int, hostIP string) *corev1.Pod {
		pod := workerPod(ordinal, hostIP)
		now := metav1.Now()
		pod.DeletionTimestamp = &now
		// A finalizer is required for the fake client to accept a
		// DeletionTimestamp on Create - a real API server sets one
		// automatically only in response to a Delete call.
		pod.Finalizers = []string{"fluid.io/test-finalizer"}
		return pod
	}

	Context("when the pod targeted for removal is already gone", func() {
		It("treats a NotFound pod as already decommissioned", func() {
			engine = newEngineWithPods()
			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
		})
	})

	Context("when the pod targeted for removal is already terminating", func() {
		It("skips it instead of re-issuing a decommission for an already-handled worker", func() {
			// Regression test: workers.Status.Replicas (used to decide
			// whether to enter the decommission block at all) stays elevated
			// until a terminating pod is fully gone, which can briefly
			// re-trigger this scan for a pod from an earlier, already
			// successful drain. Kubernetes has no distinct "Terminating"
			// Phase - the pod keeps reporting Running during its grace
			// period - so DeletionTimestamp is the only reliable signal.
			engine = newEngineWithPods(terminatingWorkerPod(1, "10.0.0.1"))
			decommissionCalled := false
			patch := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error {
					decommissionCalled = true
					return nil
				})
			defer patch.Reset()

			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
			Expect(decommissionCalled).To(BeFalse())
		})
	})

	Context("when the pod's HostIP is IPv6", func() {
		It("brackets it in the decommission address", func() {
			engine = newEngineWithPods(workerPod(1, "2001:db8::1"))
			var capturedAddrs []string
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, addrs []string) error {
					capturedAddrs = append([]string(nil), addrs...)
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) { return 1, nil })
			defer patch2.Reset()

			_, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(capturedAddrs).To(Equal([]string{"[2001:db8::1]:" + fmt.Sprint(defaultWorkerWebPort)}))
		})
	})

	Context("when the pod's actual web port differs from the spec/default value", func() {
		It("targets the port the alluxio-worker container actually binds, not the spec/default port", func() {
			// Regression test: in host network mode (the default), a worker
			// web port that isn't pinned via spec.worker.ports.web is
			// allocated dynamically by the operator's port allocator and
			// only ever written into the worker container's own spec, never
			// back into runtime.Spec.Worker.Ports. Targeting
			// getWorkerWebPort's spec/default value instead of the pod's
			// actual bound port would silently address the wrong worker.
			engine = newEngineWithPods(workerPodWithWebPort(1, "10.0.0.1", 20493))
			var capturedAddrs []string
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, addrs []string) error {
					capturedAddrs = append([]string(nil), addrs...)
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) { return 1, nil })
			defer patch2.Reset()

			drained, addrs, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
			Expect(addrs).To(Equal([]string{"10.0.0.1:20493"}))
			Expect(capturedAddrs).To(Equal([]string{"10.0.0.1:20493"}))
		})
	})

	Context("when the pod has not yet been assigned a host IP", func() {
		It("skips the pod and treats it as having nothing to drain", func() {
			engine = newEngineWithPods(workerPod(1, ""))
			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
		})

		It("still decommissions a sibling pod that does have a host IP", func() {
			engine = newEngineWithPods(workerPod(1, ""), workerPod(2, "10.0.0.1"))

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

			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
			Expect(capturedAddrs).To(Equal([]string{"10.0.0.1:" + fmt.Sprint(defaultWorkerWebPort)}))
		})
	})

	Context("when a decommission attempt is already tracked", func() {
		It("does not re-issue the decommission command, only polls active worker count", func() {
			targetAddr := "10.0.0.1:" + fmt.Sprint(defaultWorkerWebPort)
			rt.Status.Conditions = []v1alpha1.RuntimeCondition{
				utils.NewRuntimeCondition(v1alpha1.RuntimeWorkerDecommissioning,
					v1alpha1.RuntimeWorkerDecommissioningReason, decommissionTargetsMessage([]string{targetAddr}), corev1.ConditionTrue),
			}
			engine = newEngineWithPods(workerPod(1, "10.0.0.1"))

			decommissionCalled := false
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error {
					decommissionCalled = true
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) {
					return 1, nil
				})
			defer patch2.Reset()

			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
			Expect(decommissionCalled).To(BeFalse())
		})
	})

	Context("when the target set has widened since the tracked attempt", func() {
		It("retries instead of treating the stale narrower attempt as done", func() {
			// Regression test: the user scaled down further (e.g. 3 -> 2 -> 1)
			// before the first attempt (targeting only worker-2) finished.
			// The tracked condition still has the success Reason from that
			// first attempt, but its message only lists worker-2 - not
			// worker-1, which drainScalingDownWorkers now also targets. A
			// Reason-only check would treat this as already handled and
			// leave worker-1 never actually told to decommission.
			workerTwoAddr := "10.0.0.2:" + fmt.Sprint(defaultWorkerWebPort)
			rt.Status.Conditions = []v1alpha1.RuntimeCondition{
				utils.NewRuntimeCondition(v1alpha1.RuntimeWorkerDecommissioning,
					v1alpha1.RuntimeWorkerDecommissioningReason, decommissionTargetsMessage([]string{workerTwoAddr}), corev1.ConditionTrue),
			}
			engine = newEngineWithPods(workerPod(1, "10.0.0.1"), workerPod(2, "10.0.0.2"))

			var capturedAddrs []string
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, addrs []string) error {
					capturedAddrs = append([]string(nil), addrs...)
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) { return 1, nil })
			defer patch2.Reset()

			// desiredReplicas=1, currentReplicas=3: targets worker-1 and
			// worker-2, wider than the tracked attempt's worker-2-only set.
			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
			Expect(capturedAddrs).To(ConsistOf("10.0.0.1:"+fmt.Sprint(defaultWorkerWebPort), workerTwoAddr))
		})
	})

	Context("when a previously tracked decommission attempt failed", func() {
		It("retries the decommission command instead of skipping it", func() {
			rt.Status.Conditions = []v1alpha1.RuntimeCondition{
				utils.NewRuntimeCondition(v1alpha1.RuntimeWorkerDecommissioning,
					v1alpha1.RuntimeWorkerDecommissionFailedReason, "failed earlier", corev1.ConditionTrue),
			}
			engine = newEngineWithPods(workerPod(1, "10.0.0.1"))

			decommissionCalled := false
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error {
					decommissionCalled = true
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) {
					return 1, nil
				})
			defer patch2.Reset()

			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
			Expect(decommissionCalled).To(BeTrue())
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

			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
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

			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
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

			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 2)
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

			drained, _, err := engine.drainScalingDownWorkers(context.TODO(), rt, 1, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(drained).To(BeTrue())
			Expect(capturedAddrs).To(HaveLen(1))
		})
	})
})

var _ = Describe("AlluxioEngine getWorkerWebPort", Label("pkg.ddc.alluxio.replicas_drain_test.go"), func() {
	var engine *AlluxioEngine

	BeforeEach(func() {
		engine = newAlluxioEngineREP(fake.NewFakeClientWithScheme(testScheme), testDrainWorkerSts, testDrainNamespace)
	})

	It("returns the configured web port when set", func() {
		rt := &v1alpha1.AlluxioRuntime{
			Spec: v1alpha1.AlluxioRuntimeSpec{
				Worker: v1alpha1.AlluxioCompTemplateSpec{
					Ports: map[string]int{"web": 12345},
				},
			},
		}
		Expect(engine.getWorkerWebPort(rt)).To(Equal(12345))
	})

	It("falls back to the default port when unset", func() {
		rt := &v1alpha1.AlluxioRuntime{}
		Expect(engine.getWorkerWebPort(rt)).To(Equal(defaultWorkerWebPort))
	})

	It("falls back to the default port when the configured value is not positive", func() {
		rt := &v1alpha1.AlluxioRuntime{
			Spec: v1alpha1.AlluxioRuntimeSpec{
				Worker: v1alpha1.AlluxioCompTemplateSpec{
					Ports: map[string]int{"web": 0},
				},
			},
		}
		Expect(engine.getWorkerWebPort(rt)).To(Equal(defaultWorkerWebPort))
	})
})

var _ = Describe("workerWebPortFromPod", Label("pkg.ddc.alluxio.replicas_drain_test.go"), func() {
	It("returns the port bound by the alluxio-worker container's \"web\" port", func() {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "alluxio-worker", Ports: []corev1.ContainerPort{{Name: "web", ContainerPort: 20493}}},
				},
			},
		}
		port, ok := workerWebPortFromPod(pod)
		Expect(ok).To(BeTrue())
		Expect(port).To(Equal(20493))
	})

	It("ignores the alluxio-job-worker container's own \"web\" port", func() {
		// alluxio-job-worker also exposes a differently-numbered port named
		// "web" (see charts/alluxio/templates/worker/statefulset.yaml) - it
		// must not be mistaken for the alluxio-worker container's web port.
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "alluxio-job-worker", Ports: []corev1.ContainerPort{{Name: "web", ContainerPort: 30003}}},
				},
			},
		}
		_, ok := workerWebPortFromPod(pod)
		Expect(ok).To(BeFalse())
	})

	It("returns false when the alluxio-worker container has no \"web\" port", func() {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "alluxio-worker"},
				},
			},
		}
		_, ok := workerWebPortFromPod(pod)
		Expect(ok).To(BeFalse())
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
			Status:     corev1.PodStatus{HostIP: "10.0.0.5", Phase: corev1.PodRunning},
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

	Context("when scaling down to zero replicas", func() {
		It("skips graceful decommission entirely and scales down directly", func() {
			// There are no surviving workers left to redistribute cached
			// blocks to once every worker is being removed, so graceful
			// decommission has nothing to accomplish here - only a
			// guaranteed wait for the deadline before falling back to the
			// same ungraceful scale-down anyway.
			rt := &v1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime, Namespace: deadlineTestNs},
				Spec:       v1alpha1.AlluxioRuntimeSpec{Replicas: 0},
				Status:     v1alpha1.RuntimeStatus{DesiredWorkerNumberScheduled: 1},
			}
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime + "-worker", Namespace: deadlineTestNs},
				Spec:       appsv1.StatefulSetSpec{Replicas: ptr.To[int32](1)},
				Status:     appsv1.StatefulSetStatus{Replicas: 1},
			}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime + "-worker-0", Namespace: deadlineTestNs},
				Status:     corev1.PodStatus{HostIP: "10.0.0.5", Phase: corev1.PodRunning},
			}
			dataset := &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime, Namespace: deadlineTestNs},
			}
			fakeClient := fake.NewFakeClientWithScheme(testScheme, rt, sts, pod, dataset)
			engine := newAlluxioEngineREP(fakeClient, deadlineTestRuntime, deadlineTestNs)

			decommissionCalled := false
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error {
					decommissionCalled = true
					return nil
				})
			defer patch1.Reset()
			countCalled := false
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) {
					countCalled = true
					return 0, nil
				})
			defer patch2.Reset()

			err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
				Log: fake.NullLogger(), Recorder: record.NewFakeRecorder(300),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(decommissionCalled).To(BeFalse())
			Expect(countCalled).To(BeFalse())

			cond := getCondition(engine)
			Expect(cond).To(BeNil())

			var updatedSts appsv1.StatefulSet
			Expect(engine.Client.Get(context.TODO(),
				types.NamespacedName{Name: deadlineTestRuntime + "-worker", Namespace: deadlineTestNs}, &updatedSts)).To(Succeed())
			Expect(*updatedSts.Spec.Replicas).To(Equal(int32(0)))
		})

		It("clears a lingering condition from an earlier scale-in that a scale-to-zero cut short, so a later scale-down starts fresh", func() {
			// Regression test for the exact sequence cheyang traced: 3 -> 1
			// tracks a decommission condition (Status=True); before that
			// drain finishes, the user scales straight to 0. The block above
			// is skipped entirely (Replicas()==0), so without this clearing
			// step the condition would be stranded at Status=True forever.
			// A later, unrelated scale-down would then read its stale
			// LastTransitionTime as its own decommissionStart and find
			// itself already past defaultWorkerDecommissionDeadline on its
			// very first reconcile - forcing scale-down through immediately
			// instead of attempting a graceful drain at all.
			staleCond := utils.NewRuntimeCondition(v1alpha1.RuntimeWorkerDecommissioning,
				v1alpha1.RuntimeWorkerDecommissioningReason, "abandoned 3 -> 1 attempt", corev1.ConditionTrue)
			staleCond.LastTransitionTime = metav1.NewTime(time.Now().Add(-defaultWorkerDecommissionDeadline - time.Hour))

			rt := &v1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime, Namespace: deadlineTestNs},
				Spec:       v1alpha1.AlluxioRuntimeSpec{Replicas: 0},
				Status: v1alpha1.RuntimeStatus{
					DesiredWorkerNumberScheduled: 1,
					Conditions:                   []v1alpha1.RuntimeCondition{staleCond},
				},
			}
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime + "-worker", Namespace: deadlineTestNs},
				Spec:       appsv1.StatefulSetSpec{Replicas: ptr.To[int32](1)},
				Status:     appsv1.StatefulSetStatus{Replicas: 1},
			}
			dataset := &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: deadlineTestRuntime, Namespace: deadlineTestNs},
			}
			fakeClient := fake.NewFakeClientWithScheme(testScheme, rt, sts, dataset)
			engine := newAlluxioEngineREP(fakeClient, deadlineTestRuntime, deadlineTestNs)

			// Step 1: reconcile the scale-to-zero. This must clear the
			// stranded condition rather than just skip past it.
			err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
				Log: fake.NullLogger(), Recorder: record.NewFakeRecorder(300),
			})
			Expect(err).NotTo(HaveOccurred())

			clearedCond := getCondition(engine)
			Expect(clearedCond).NotTo(BeNil())
			Expect(clearedCond.Status).To(Equal(corev1.ConditionFalse))

			// Step 2: time passes; the runtime is scaled back up, then a
			// genuine new scale-down (3 -> 1) is requested. If the earlier
			// condition had leaked through, this reconcile would immediately
			// force scale-down to proceed (it'd compute elapsed against the
			// ancient stale timestamp, already past the deadline) instead of
			// starting a real graceful drain.
			var currentSts appsv1.StatefulSet
			Expect(engine.Client.Get(context.TODO(),
				types.NamespacedName{Name: deadlineTestRuntime + "-worker", Namespace: deadlineTestNs}, &currentSts)).To(Succeed())
			currentSts.Spec.Replicas = ptr.To[int32](3)
			Expect(engine.Client.Update(context.TODO(), &currentSts)).To(Succeed())
			// Spec and Status are separate subresources on this fake client
			// (see fake.NewFakeClientWithScheme's WithStatusSubresource) - a
			// plain Update only persists the Spec change above.
			currentSts.Status.Replicas = 3
			Expect(engine.Client.Status().Update(context.TODO(), &currentSts)).To(Succeed())

			for _, ord := range []int{0, 1, 2} {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-worker-%d", deadlineTestRuntime, ord), Namespace: deadlineTestNs},
					Status:     corev1.PodStatus{HostIP: fmt.Sprintf("10.0.0.%d", ord+10), Phase: corev1.PodRunning},
				}
				Expect(engine.Client.Create(context.TODO(), pod)).To(Succeed())
			}

			var newRt v1alpha1.AlluxioRuntime
			Expect(engine.Client.Get(context.TODO(),
				types.NamespacedName{Name: deadlineTestRuntime, Namespace: deadlineTestNs}, &newRt)).To(Succeed())
			newRt.Spec.Replicas = 1
			Expect(engine.Client.Update(context.TODO(), &newRt)).To(Succeed())

			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error { return nil })
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) { return 3, nil })
			defer patch2.Reset()

			err = engine.SyncReplicas(cruntime.ReconcileRequestContext{
				Log: fake.NullLogger(), Recorder: record.NewFakeRecorder(300),
			})
			// A fresh attempt that hasn't finished within one reconcile
			// reports errWorkersNotYetDrained and keeps requeuing - not a
			// forced-through proceed, which is what a leaked stale
			// decommissionStart would have produced instead.
			Expect(errors.Is(err, errWorkersNotYetDrained)).To(BeTrue())

			freshCond := getCondition(engine)
			Expect(freshCond).NotTo(BeNil())
			Expect(freshCond.Status).To(Equal(corev1.ConditionTrue))
			Expect(time.Since(freshCond.LastTransitionTime.Time)).To(BeNumerically("<", time.Minute))
		})
	})

	Context("when a new attempt targets the exact same addresses as a previously cleared one", func() {
		It("flips the condition back to True instead of leaving it False", func() {
			// Regression test: clearDecommissioningCondition only flips
			// Status to False - it leaves Reason/Message as they were. If a
			// later decommission attempt targets the identical address set
			// (e.g. scaled back down to the same replica count) and also
			// reaches the master (Reason unchanged), Reason and Message
			// alone can't tell this new, still-active attempt apart from the
			// old, finished one - only Status differs.
			clearedCond := utils.NewRuntimeCondition(v1alpha1.RuntimeWorkerDecommissioning,
				v1alpha1.RuntimeWorkerDecommissioningReason,
				decommissionTargetsMessage([]string{"10.0.0.5:" + fmt.Sprint(defaultWorkerWebPort)}), corev1.ConditionFalse)
			engine := newFixtures(&clearedCond)

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
		})
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

	Context("when DecommissionWorkers itself fails (e.g. unsupported on the master's Alluxio version)", func() {
		It("still records a decommission-start marker instead of erroring out with nothing tracked", func() {
			engine := newFixtures(nil)
			decommissionErr := errors.New("decommissionWorker: command not found")
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error { return decommissionErr })
			defer patch1.Reset()

			err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
				Log: fake.NullLogger(), Recorder: record.NewFakeRecorder(300),
			})
			// The real error is returned (and logged at Error level by the
			// caller) rather than masked as the normal/expected
			// errWorkersNotYetDrained, so it's visible on every reconcile.
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, errWorkersNotYetDrained)).To(BeFalse())
			Expect(err.Error()).To(ContainSubstring("command not found"))

			cond := getCondition(engine)
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			Expect(cond.Reason).To(Equal(v1alpha1.RuntimeWorkerDecommissionFailedReason))
		})

		It("retries on the next reconcile instead of treating the failed attempt as done", func() {
			// Regression test: a failed attempt must not be recorded the same
			// way as a successful one, or drainScalingDownWorkers would skip
			// ever retrying DecommissionWorkers and just poll a worker count
			// that can never drop, since nothing was ever actually told to
			// decommission.
			engine := newFixtures(nil)
			decommissionErr := errors.New("transient: connection reset")
			callCount := 0
			patch1 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.DecommissionWorkers,
				func(_ operations.AlluxioFileUtils, _ []string) error {
					callCount++
					if callCount == 1 {
						return decommissionErr
					}
					return nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(operations.AlluxioFileUtils.CountActiveWorkers,
				func(_ operations.AlluxioFileUtils) (int, error) { return 1, nil })
			defer patch2.Reset()

			reconcile := func() error {
				return engine.SyncReplicas(cruntime.ReconcileRequestContext{
					Log: fake.NullLogger(), Recorder: record.NewFakeRecorder(300),
				})
			}

			firstErr := reconcile()
			Expect(firstErr).To(HaveOccurred())
			Expect(errors.Is(firstErr, errWorkersNotYetDrained)).To(BeFalse())
			cond := getCondition(engine)
			Expect(cond.Reason).To(Equal(v1alpha1.RuntimeWorkerDecommissionFailedReason))

			secondErr := reconcile()
			Expect(callCount).To(Equal(2), "the second reconcile must retry DecommissionWorkers, not skip it")
			Expect(secondErr).NotTo(HaveOccurred())

			finalCond := getCondition(engine)
			Expect(finalCond).NotTo(BeNil())
			Expect(finalCond.Status).To(Equal(corev1.ConditionFalse))
		})

		It("still forces the scale-down to proceed past the deadline instead of stalling forever", func() {
			// The pre-seeded condition's message deliberately doesn't match
			// the current target set (see decommissionTargetsMessage), so
			// decommissionSucceeded treats this attempt as stale and
			// DecommissionWorkers gets retried rather than skipped - it just
			// doesn't matter to the outcome, since CountActiveWorkers keeps
			// reporting the worker as active either way and the deadline
			// forces the scale-down through regardless.
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
