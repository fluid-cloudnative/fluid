package ctrl

import (
	"context"

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

	Describe("Test Helper.CheckAndUpdateMasterStatus()", func() {
		When("master sts is ready", func() {
			BeforeEach(func() {
				masterSts.Spec.Replicas = ptr.To[int32](1)
				masterSts.Status.AvailableReplicas = 1
				masterSts.Status.Replicas = 1
				masterSts.Status.ReadyReplicas = 1
			})

			When("applying to AlluxioRuntime", func() {
				var alluxioruntime *datav1alpha1.AlluxioRuntime
				BeforeEach(func() {
					alluxioruntime = &datav1alpha1.AlluxioRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-alluxio",
							Namespace: "fluid",
						},
						Spec:   datav1alpha1.AlluxioRuntimeSpec{},
						Status: datav1alpha1.RuntimeStatus{},
					}
					resources = append(resources, alluxioruntime)
				})
				It("should update AlluxioRuntime's status successfully", func() {
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
					Expect(gotRuntime.Status.DesiredMasterNumberScheduled).To(Equal(*masterSts.Spec.Replicas))
					Expect(gotRuntime.Status.CurrentMasterNumberScheduled).To(Equal(masterSts.Status.Replicas))
					Expect(gotRuntime.Status.MasterNumberReady).To(Equal(masterSts.Status.ReadyReplicas))

					Expect(gotRuntime.Status.Conditions).To(HaveLen(1))
					Expect(gotRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
					Expect(gotRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				})
			})
		})

	})

})
