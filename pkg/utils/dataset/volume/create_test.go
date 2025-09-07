package volume

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create Volume Tests", Label("pkg.utils.dataset.volume.create_test.go"), func() {
	var (
		scheme      *runtime.Scheme
		client      client.Client
		runtimeInfo base.RuntimeInfoInterface
		dataset     *datav1alpha1.Dataset
		daemonset   *appsv1.DaemonSet
		resources   []runtime.Object
		log         logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = v1.AddToScheme(scheme)
		_ = appsv1.AddToScheme(scheme)
		_ = datav1alpha1.AddToScheme(scheme)

		var err error
		runtimeInfo, err = base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
		Expect(err).To(BeNil())

		dataset = &datav1alpha1.Dataset{ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"}}

		daemonset = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-fuse",
				Namespace: "fluid",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "fuse:v1",
							},
						},
					},
				},
			},
		}

		resources = []runtime.Object{
			dataset,
			daemonset,
		}

		log = fake.NullLogger()
	})

	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(scheme, resources...)
	})

	Context("Test CreatePersistentVolumeForRuntime()", func() {
		When("runtime info has defined node affinity on it", func() {
			BeforeEach(func() {
				runtimeInfo.SetFuseNodeSelector(map[string]string{"test-affinity": "true"})
			})
			It("should create PV with node affinity and annotations", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				Expect(list.Items).To(HaveLen(1))
				pv := list.Items[0]
				Expect(pv.Spec.CSI).NotTo(BeNil())
				Expect(pv.Spec.CSI.VolumeAttributes).To(HaveKeyWithValue(common.VolumeAttrFluidPath, "/mnt"))
				Expect(pv.Spec.NodeAffinity).NotTo(BeNil())
			})
		})

		When("PV with fluid annotations is already exists", func() {
			BeforeEach(func() {
				// Pre-create PV with expected annotations
				pv := &v1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{
						Name:        runtimeInfo.GetPersistentVolumeName(),
						Annotations: common.GetExpectedFluidAnnotations(),
					},
				}

				resources = append(resources, pv)
			})
			It("should skip creating PV if it already exists with fluid annotations", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				Expect(list.Items).To(HaveLen(1))
			})
		})
	})

	Context("Test CreatePersistentVolumeClaimForRuntime()", func() {
		It("should create PVC and record fuse generation if exists", func() {
			runtimeInfo.SetFuseName("hbase-fuse")
			Expect(CreatePersistentVolumeClaimForRuntime(client, runtimeInfo, log)).To(Succeed())
			var list v1.PersistentVolumeClaimList
			Expect(client.List(context.TODO(), &list)).To(Succeed())
			Expect(list.Items).ToNot(BeEmpty())
		})

		It("should create PVC even if no fuse daemonset found (no generation label)", func() {
			Expect(CreatePersistentVolumeClaimForRuntime(client, runtimeInfo, log)).To(Succeed())
			var list v1.PersistentVolumeClaimList
			Expect(client.List(context.TODO(), &list)).To(Succeed())
			Expect(list.Items).ToNot(BeEmpty())
		})
	})
})
