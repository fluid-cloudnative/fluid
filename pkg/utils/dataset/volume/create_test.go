package volume

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
				Expect(pv.Labels).To(HaveKeyWithValue(runtimeInfo.GetCommonLabelName(), "true"))
				Expect(pv.Labels).To(HaveKeyWithValue(common.LabelAnnotationDatasetId, "fluid-hbase"))
				Expect(pv.Spec.StorageClassName).To(Equal(common.FluidStorageClass))
				Expect(pv.Spec.CSI).NotTo(BeNil())
				Expect(pv.Spec.CSI.VolumeAttributes).To(HaveKeyWithValue(common.VolumeAttrFluidPath, "/mnt"))
				Expect(pv.Spec.NodeAffinity.Required.NodeSelectorTerms).To(HaveLen(1))
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

		When("Related Dataset has set a explicit access mode", func() {
			BeforeEach(func() {
				dataset.Spec.AccessModes = append(dataset.Spec.AccessModes, v1.ReadWriteMany)
			})

			It("should create a pv with same access mode", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				Expect(list.Items).To(HaveLen(1))
				pv := list.Items[0]
				Expect(pv.Spec.AccessModes).To(HaveLen(1))
				Expect(pv.Spec.AccessModes).To(ContainElement(v1.ReadWriteMany))
			})
		})

		When("dataset has no explicit access mode", func() {
			It("should create pv with default access mode", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				Expect(list.Items).To(HaveLen(1))
				pv := list.Items[0]
				Expect(pv.Spec.AccessModes).To(HaveLen(1))
				Expect(pv.Spec.AccessModes).To(ContainElement(v1.ReadOnlyMany))
			})
		})

		When("dataset has pvc storage capacity annotation set", func() {
			BeforeEach(func() {
				if dataset.Annotations == nil {
					dataset.Annotations = map[string]string{}
				}
				dataset.Annotations[utils.PVCStorageAnnotation] = "10Gi"
			})
			It("should use annotated storage capacity", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				pv := list.Items[0]
				q := pv.Spec.Capacity[v1.ResourceStorage]
				expected := resource.MustParse("10Gi")
				Expect(q.Cmp(expected)).To(Equal(0))
			})
		})

		When("dataset pvc storage capacity annotation is invalid", func() {
			BeforeEach(func() {
				if dataset.Annotations == nil {
					dataset.Annotations = map[string]string{}
				}
				dataset.Annotations[utils.PVCStorageAnnotation] = "invalid-size"
			})
			It("should fallback to default storage capacity", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				pv := list.Items[0]
				q := pv.Spec.Capacity[v1.ResourceStorage]
				expected := resource.MustParse(utils.DefaultStorageCapacity)
				Expect(q.Cmp(expected)).To(Equal(0))
			})
		})

		When("creating pv csi attributes", func() {
			It("should set mountType/namespace/name and claimRef", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				pv := list.Items[0]
				Expect(pv.Spec.CSI).NotTo(BeNil())
				attrs := pv.Spec.CSI.VolumeAttributes
				Expect(attrs).To(HaveKeyWithValue(common.VolumeAttrMountType, "alluxio"))
				Expect(attrs).To(HaveKeyWithValue(common.VolumeAttrNamespace, runtimeInfo.GetNamespace()))
				Expect(attrs).To(HaveKeyWithValue(common.VolumeAttrName, runtimeInfo.GetName()))
				Expect(pv.Spec.ClaimRef).NotTo(BeNil())
				Expect(pv.Spec.ClaimRef.Name).To(Equal(runtimeInfo.GetName()))
				Expect(pv.Spec.ClaimRef.Namespace).To(Equal(runtimeInfo.GetNamespace()))
			})
		})

		When("runtime annotations contain skip-check-mount-ready target", func() {
			BeforeEach(func() {
				var err error
				runtimeInfo, err = base.BuildRuntimeInfo("hbase", "fluid", "alluxio", base.WithAnnotations(map[string]string{
					common.AnnotationSkipCheckMountReadyTarget: "All",
				}))
				Expect(err).To(BeNil())
			})
			It("should propagate to pv csi volumeAttributes", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				pv := list.Items[0]
				Expect(pv.Spec.CSI.VolumeAttributes).To(HaveKeyWithValue(common.AnnotationSkipCheckMountReadyTarget, "All"))
			})
		})

		When("metadataList selects PV with labels/annotations and symlink method", func() {
			BeforeEach(func() {
				meta := []datav1alpha1.Metadata{{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels: map[string]string{
							common.LabelNodePublishMethod: common.NodePublishMethodSymlink,
							"x-extra":                     "y",
						},
						Annotations: map[string]string{
							"a1": "b1",
						},
					},
					Selector: metav1.GroupKind{Group: v1.GroupName, Kind: "PersistentVolume"},
				}}
				var err error
				runtimeInfo, err = base.BuildRuntimeInfo("hbase", "fluid", "alluxio", base.WithMetadataList(meta))
				Expect(err).To(BeNil())
			})
			It("should merge into pv and set node_publish_method", func() {
				Expect(CreatePersistentVolumeForRuntime(client, runtimeInfo, "/mnt", "alluxio", log)).To(Succeed())
				var list v1.PersistentVolumeList
				Expect(client.List(context.TODO(), &list)).To(Succeed())
				pv := list.Items[0]
				Expect(pv.Labels).To(HaveKeyWithValue("x-extra", "y"))
				Expect(pv.Annotations).To(HaveKeyWithValue("a1", "b1"))
				Expect(pv.Spec.CSI.VolumeAttributes).To(HaveKeyWithValue(common.NodePublishMethod, common.NodePublishMethodSymlink))
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
