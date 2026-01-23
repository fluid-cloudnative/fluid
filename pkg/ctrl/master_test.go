package ctrl

import (
	"context"
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

var _ = Describe("Ctrl Master Tests", Label("pkg.ctrl.master_test.go"), func() {
	var helper *Helper
	var resources []runtime.Object
	var masterSts *appsv1.StatefulSet
	var k8sClient client.Client
	var runtimeInfo base.RuntimeInfoInterface

	BeforeEach(func() {
		masterSts = mockRuntimeStatefulset("test-helper", "fluid")
		resources = []runtime.Object{
			masterSts,
		}
	})

	JustBeforeEach(func() {
		k8sClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		helper = BuildHelper(runtimeInfo, k8sClient, fake.NullLogger())
	})

	Describe("Test Helper.CheckAndSyncMasterStatus()", func() {
		Context("Error handling - StatefulSet not found", func() {
			It("should return error when StatefulSet doesn't exist", func() {
				alluxioruntime := &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{},
				}
				resources = []runtime.Object{alluxioruntime}
				k8sClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
				helper = BuildHelper(runtimeInfo, k8sClient, fake.NullLogger())

				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.AlluxioRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: "fluid", Name: "non-existent-sts"})
				Expect(err).NotTo(BeNil())
				Expect(ready).To(BeFalse())
			})
		})

		Context("Error handling - getRuntimeFn returns error", func() {
			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](1)
				masterSts.Status.Replicas = 1
				masterSts.Status.ReadyReplicas = 1
			})

			It("should return error when getRuntimeFn fails", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					return nil, fmt.Errorf("failed to get runtime")
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: masterSts.Namespace, Name: masterSts.Name})
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to update master ready status in runtime status"))
				Expect(ready).To(BeFalse())
			})
		})

		Context("StatefulSet with nil Replicas", func() {
			var alluxioruntime *datav1alpha1.AlluxioRuntime

			BeforeEach(func() {
				masterSts.Spec.Replicas = nil
				masterSts.Status.Replicas = 0
				masterSts.Status.ReadyReplicas = 0

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{},
				}
				resources = append(resources, alluxioruntime)
			})

		})

		Context("Master phase: Ready", func() {
			var alluxioruntime *datav1alpha1.AlluxioRuntime

			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](3)
				masterSts.Status.AvailableReplicas = 3
				masterSts.Status.Replicas = 3
				masterSts.Status.ReadyReplicas = 3

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{},
				}
				resources = append(resources, alluxioruntime)
			})

			It("should set phase to Ready and update conditions", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.AlluxioRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: masterSts.Namespace, Name: masterSts.Name})
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				gotRuntime := &datav1alpha1.AlluxioRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseReady))
				Expect(gotRuntime.Status.DesiredMasterNumberScheduled).To(Equal(int32(3)))
				Expect(gotRuntime.Status.CurrentMasterNumberScheduled).To(Equal(int32(3)))
				Expect(gotRuntime.Status.MasterNumberReady).To(Equal(int32(3)))

				Expect(gotRuntime.Status.Conditions).To(HaveLen(2))
				Expect(gotRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeMasterInitialized))
				Expect(gotRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				Expect(gotRuntime.Status.Conditions[0].Reason).To(Equal(datav1alpha1.RuntimeMasterInitializedReason))
				Expect(gotRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
				Expect(gotRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionTrue))
				Expect(gotRuntime.Status.Conditions[1].Reason).To(Equal(datav1alpha1.RuntimeMasterReadyReason))
				Expect(gotRuntime.Status.Conditions[1].Message).To(Equal("The master is ready."))
			})
		})

		Context("Master phase: PartialReady", func() {
			var alluxioruntime *datav1alpha1.AlluxioRuntime

			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](3)
				masterSts.Status.AvailableReplicas = 2
				masterSts.Status.Replicas = 3
				masterSts.Status.ReadyReplicas = 2

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{},
				}
				resources = append(resources, alluxioruntime)
			})

			It("should set phase to PartialReady and update conditions", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.AlluxioRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: masterSts.Namespace, Name: masterSts.Name})
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				gotRuntime := &datav1alpha1.AlluxioRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhasePartialReady))
				Expect(gotRuntime.Status.MasterNumberReady).To(Equal(int32(2)))

				Expect(gotRuntime.Status.Conditions).To(HaveLen(2))
				Expect(gotRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
				Expect(gotRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionTrue))
				Expect(gotRuntime.Status.Conditions[1].Message).To(Equal("The master is partially ready."))
			})
		})

		Context("Master phase: NotReady", func() {
			var alluxioruntime *datav1alpha1.AlluxioRuntime

			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](3)
				masterSts.Status.AvailableReplicas = 0
				masterSts.Status.Replicas = 0
				masterSts.Status.ReadyReplicas = 0

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{},
				}
				resources = append(resources, alluxioruntime)
			})

			It("should set phase to NotReady and update conditions", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.AlluxioRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: masterSts.Namespace, Name: masterSts.Name})
				Expect(err).To(BeNil())
				Expect(ready).To(BeFalse())

				gotRuntime := &datav1alpha1.AlluxioRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(gotRuntime.Status.MasterNumberReady).To(Equal(int32(0)))

				Expect(gotRuntime.Status.Conditions).To(HaveLen(2))
				Expect(gotRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
				Expect(gotRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionFalse))
				Expect(gotRuntime.Status.Conditions[1].Message).To(Equal("The master is not ready."))
			})
		})

		Context("Status with existing conditions", func() {
			var alluxioruntime *datav1alpha1.AlluxioRuntime

			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](1)
				masterSts.Status.Replicas = 1
				masterSts.Status.ReadyReplicas = 1

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{
						Conditions: []datav1alpha1.RuntimeCondition{
							{
								Type:               datav1alpha1.RuntimeMasterInitialized,
								Status:             corev1.ConditionTrue,
								Reason:             datav1alpha1.RuntimeMasterInitializedReason,
								Message:            "Already initialized",
								LastProbeTime:      metav1.Now(),
								LastTransitionTime: metav1.Now(),
							},
						},
					},
				}
				resources = append(resources, alluxioruntime)
			})

			It("should update existing conditions correctly", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.AlluxioRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: masterSts.Namespace, Name: masterSts.Name})
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				gotRuntime := &datav1alpha1.AlluxioRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.Conditions).To(HaveLen(2))
			})
		})

		Context("No status update needed (idempotency)", func() {
			var alluxioruntime *datav1alpha1.AlluxioRuntime

			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](1)
				masterSts.Status.Replicas = 1
				masterSts.Status.ReadyReplicas = 1

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{
						DesiredMasterNumberScheduled: 1,
						CurrentMasterNumberScheduled: 1,
						MasterNumberReady:            1,
						MasterPhase:                  datav1alpha1.RuntimePhaseReady,
						Conditions: []datav1alpha1.RuntimeCondition{
							{
								Type:    datav1alpha1.RuntimeMasterInitialized,
								Status:  corev1.ConditionTrue,
								Reason:  datav1alpha1.RuntimeMasterInitializedReason,
								Message: "The master is initialized.",
							},
							{
								Type:    datav1alpha1.RuntimeMasterReady,
								Status:  corev1.ConditionTrue,
								Reason:  datav1alpha1.RuntimeMasterReadyReason,
								Message: "The master is ready.",
							},
						},
					},
				}
				resources = append(resources, alluxioruntime)
			})

			It("should skip update when status hasn't changed", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.AlluxioRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: masterSts.Namespace, Name: masterSts.Name})
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})

		Context("JindoRuntime support", func() {
			var jindoruntime *datav1alpha1.JindoRuntime

			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](2)
				masterSts.Status.Replicas = 2
				masterSts.Status.ReadyReplicas = 2

				jindoruntime = &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-jindo",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{},
				}
				resources = append(resources, jindoruntime)
			})

			It("should work with JindoRuntime", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.JindoRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: jindoruntime.Namespace, Name: jindoruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: masterSts.Namespace, Name: masterSts.Name})
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				gotRuntime := &datav1alpha1.JindoRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: jindoruntime.Namespace, Name: jindoruntime.Name}, gotRuntime)
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseReady))
			})
		})

		Context("JuiceFSRuntime support", func() {
			var juicefsruntime *datav1alpha1.JuiceFSRuntime

			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](1)
				masterSts.Status.Replicas = 0
				masterSts.Status.ReadyReplicas = 0

				juicefsruntime = &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-juicefs",
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{},
				}
				resources = append(resources, juicefsruntime)
			})

			It("should work with JuiceFSRuntime", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.JuiceFSRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: juicefsruntime.Namespace, Name: juicefsruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: masterSts.Namespace, Name: masterSts.Name})
				Expect(err).To(BeNil())
				Expect(ready).To(BeFalse())

				gotRuntime := &datav1alpha1.JuiceFSRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: juicefsruntime.Namespace, Name: juicefsruntime.Name}, gotRuntime)
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
			})
		})
	})
})
