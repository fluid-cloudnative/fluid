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

package alluxio

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine Volume Creation Tests", Label("pkg.ddc.alluxio.create_volume_test.go"), func() {
	var (
		dataset        *datav1alpha1.Dataset
		alluxioruntime *datav1alpha1.AlluxioRuntime
		engine         *AlluxioEngine
		mockedObjects  mockedObjects
		client         client.Client
		resources      []runtime.Object
	)
	BeforeEach(func() {
		dataset, alluxioruntime = mockFluidObjectsForTests(types.NamespacedName{Namespace: "fluid", Name: "hbase"})
		engine = mockAlluxioEngineForTests(dataset, alluxioruntime)
		mockedObjects = mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
		resources = []runtime.Object{
			dataset,
			alluxioruntime,
			mockedObjects.MasterSts,
			mockedObjects.WorkerSts,
			mockedObjects.FuseDs,
		}
	})

	// JustBeforeEach is guaranteed to run after every BeforeEach()
	// So it's easy to modify resources' specs with an extra BeforeEach()
	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.runtimeInfo.SetFuseName(engine.getFuseName())
		engine.Client = client
	})

	// TestCreateVolume tests the CreateVolume function of the AlluxioEngine.
	// It verifies that the function successfully creates a PersistentVolume (PV)
	// and a PersistentVolumeClaim (PVC) for the given dataset. The test sets up
	// a fake Kubernetes client with a mock dataset and checks if exactly one PV
	// and one PVC are created after the function execution
	Describe("Test AlluxioEngine.CreateVolume()", func() {
		When("given AlluxioEngine works as expected", func() {

			It("should create volumes successfully", func() {
				err := engine.CreateVolume()
				Expect(err).To(BeNil())

				gotPVC := &corev1.PersistentVolumeClaim{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotPVC)
				Expect(err).To(BeNil())
				Expect(gotPVC.Name).To(Equal(alluxioruntime.Name))
				Expect(gotPVC.Namespace).To(Equal(alluxioruntime.Namespace))
				Expect(gotPVC.Spec.VolumeName).To(Equal(engine.runtimeInfo.GetPersistentVolumeName()))
				Expect(gotPVC.Labels[common.LabelRuntimeFuseGeneration]).To(Equal(strconv.Itoa(int(mockedObjects.FuseDs.Generation))))

				gotPv := &corev1.PersistentVolume{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: engine.runtimeInfo.GetPersistentVolumeName()}, gotPv)
				Expect(err).To(BeNil())
				Expect(gotPv.Name).To(Equal(engine.runtimeInfo.GetPersistentVolumeName()))
				Expect(gotPv.Spec.ClaimRef.Name).To(Equal(alluxioruntime.Name))
				Expect(gotPv.Spec.ClaimRef.Namespace).To(Equal(alluxioruntime.Namespace))
			})
		})
	})

	// TestCreateFusePersistentVolume tests the createFusePersistentVolume function of the AlluxioEngine.
	// It verifies that the function can successfully create a PersistentVolume (PV) for the Alluxio Fuse.
	// The test first builds the runtime information for the AlluxioEngine. Then it creates a mock dataset
	// and initializes a fake Kubernetes client with the dataset. After that, it invokes the
	// createFusePersistentVolume function of the AlluxioEngine. Finally, it checks if exactly one PV
	// is created in the fake client.
	Describe("Test AlluxioEngine.CreateFusePersistentVolume", func() {
		When("given AlluxioEngine works as expected", func() {
			It("should create fuse PV successfully", func() {
				err := engine.createFusePersistentVolume()
				Expect(err).To(BeNil())

				gotPv := &corev1.PersistentVolume{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: engine.runtimeInfo.GetPersistentVolumeName()}, gotPv)
				Expect(err).To(BeNil())
				Expect(gotPv.Name).To(Equal(engine.runtimeInfo.GetPersistentVolumeName()))
				Expect(gotPv.Labels[engine.runtimeInfo.GetCommonLabelName()]).To(Equal("true"))
				Expect(gotPv.Labels[common.LabelAnnotationDatasetId]).To(Equal(utils.GetDatasetId(engine.runtimeInfo.GetNamespace(), engine.runtimeInfo.GetName(), engine.runtimeInfo.GetOwnerDatasetUID())))
				Expect(gotPv.Annotations).To(Equal(common.GetExpectedFluidAnnotations()))
				Expect(gotPv.Spec.ClaimRef.Namespace).To(Equal(engine.namespace))
				Expect(gotPv.Spec.ClaimRef.Name).To(Equal(engine.name))
				Expect(gotPv.Spec.Capacity[corev1.ResourceStorage]).To(Equal(resource.MustParse(utils.DefaultStorageCapacity)))
				Expect(gotPv.Spec.StorageClassName).To(Equal(common.FluidStorageClass))
				Expect(gotPv.Spec.CSI.Driver).To(Equal(common.CSIDriver))
				Expect(gotPv.Spec.CSI.VolumeHandle).To(Equal(engine.runtimeInfo.GetPersistentVolumeName()))
			})
		})

		When("PersistentVolume already exists", func() {
			BeforeEach(func() {
				// fake.Client is aware of PV's namespace
				mockedObjects.PersistentVolume.Namespace = ""
				resources = append(resources, mockedObjects.PersistentVolume)
			})
			It("should not create PersistentVolumle and no error should return", func() {
				err := engine.createFusePersistentVolume()
				Expect(err).To(BeNil())
			})
		})

		When("Dataset has ReadWriteMany access mode", func() {
			BeforeEach(func() {
				dataset.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
			})
			It("should create PersistentVolumle with ReadWriteMany access mode", func() {
				err := engine.createFusePersistentVolume()
				Expect(err).To(BeNil())

				gotPv := &corev1.PersistentVolume{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: engine.runtimeInfo.GetPersistentVolumeName()}, gotPv)
				Expect(err).To(BeNil())
				Expect(gotPv.Spec.AccessModes).To(Equal([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}))
			})
		})

		When("Dataset sets storage capacity to override the default one", func() {
			BeforeEach(func() {
				dataset.Annotations = map[string]string{}
				dataset.Annotations[utils.PVCStorageAnnotation] = "30Gi"
			})
			It("should create PV with the storage capacity specified in Dataset", func() {
				err := engine.createFusePersistentVolume()
				Expect(err).To(BeNil())

				gotPv := &corev1.PersistentVolume{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: engine.runtimeInfo.GetPersistentVolumeName()}, gotPv)
				Expect(err).To(BeNil())
				Expect(gotPv.Spec.Capacity[corev1.ResourceStorage]).To(Equal(resource.MustParse("30Gi")))
			})
		})

		When("Runtime sets extra labels and annotations in metadata list", func() {
			BeforeEach(func() {
				metadata := datav1alpha1.Metadata{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels: map[string]string{
							"label1": "value1",
						},
						Annotations: map[string]string{
							"annotation1": "value2",
						},
					},
					Selector: metav1.GroupKind{
						Group: corev1.GroupName,
						Kind:  "PersistentVolume",
					},
				}
				newRuntimeInfo, err := base.BuildRuntimeInfo(alluxioruntime.Name, alluxioruntime.Namespace, common.AlluxioRuntime, base.WithMetadataList([]datav1alpha1.Metadata{metadata}))
				engine.runtimeInfo = newRuntimeInfo
				Expect(err).To(BeNil())
			})
			It("should create PV with extra labels and annotations", func() {
				err := engine.createFusePersistentVolume()
				Expect(err).To(BeNil())

				gotPv := &corev1.PersistentVolume{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: engine.runtimeInfo.GetPersistentVolumeName()}, gotPv)
				Expect(err).To(BeNil())
				Expect(gotPv.Labels).Should(HaveKeyWithValue("label1", "value1"))
				Expect(gotPv.Annotations).Should(HaveKeyWithValue("annotation1", "value2"))
			})
		})

		When("Runtime sets non-default node publish method", func() {
			BeforeEach(func() {
				metadata := datav1alpha1.Metadata{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels: map[string]string{
							common.LabelNodePublishMethod: common.NodePublishMethodSymlink,
						},
					},
					Selector: metav1.GroupKind{
						Group: corev1.GroupName,
						Kind:  "PersistentVolume",
					},
				}
				newRuntimeInfo, err := base.BuildRuntimeInfo(alluxioruntime.Name, alluxioruntime.Namespace, common.AlluxioRuntime, base.WithMetadataList([]datav1alpha1.Metadata{metadata}))
				engine.runtimeInfo = newRuntimeInfo
				Expect(err).To(BeNil())
			})
			It("should create fuse pv with specific node publish method", func() {
				err := engine.createFusePersistentVolume()
				Expect(err).To(BeNil())

				gotPv := &corev1.PersistentVolume{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: engine.runtimeInfo.GetPersistentVolumeName()}, gotPv)
				Expect(err).To(BeNil())
				Expect(gotPv.Spec.CSI.VolumeAttributes).Should(HaveKeyWithValue(common.NodePublishMethod, common.NodePublishMethodSymlink))
			})
		})
	})

	// TestCreateFusePersistentVolumeClaim tests the createFusePersistentVolumeClaim function of the AlluxioEngine.
	// It ensures that the function successfully creates a PersistentVolumeClaim (PVC) with the correct metadata.
	// The test sets up a fake Kubernetes client with a mock DaemonSet and a Dataset to simulate the environment.
	// After invoking the target function, it verifies that a PVC is created and labeled with the correct
	// fuse generation based on the input DaemonSet's generation.
	Describe("Test AlluxioEngine.CreatePersistentVolumeClaim()", func() {
		When("given AlluxioEngine works as expected", func() {
			It("should create fuse PVC successfully", func() {
				err := engine.createFusePersistentVolumeClaim()
				Expect(err).To(BeNil())

				gotPvc := &corev1.PersistentVolumeClaim{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotPvc)
				Expect(err).To(BeNil())
				Expect(gotPvc.Labels).To(HaveKeyWithValue(engine.runtimeInfo.GetCommonLabelName(), "true"))
				Expect(gotPvc.Labels).To(HaveKeyWithValue(common.LabelAnnotationDatasetId, utils.GetDatasetId(engine.runtimeInfo.GetNamespace(), engine.runtimeInfo.GetName(), engine.runtimeInfo.GetOwnerDatasetUID())))
				Expect(gotPvc.Labels).To(HaveKeyWithValue(common.LabelRuntimeFuseGeneration, strconv.Itoa(int(mockedObjects.FuseDs.Generation))))
				Expect(gotPvc.Annotations).To(Equal(common.GetExpectedFluidAnnotations()))
				Expect(gotPvc.Spec.VolumeName).To(Equal(engine.runtimeInfo.GetPersistentVolumeName()))
				Expect(*gotPvc.Spec.StorageClassName).To(Equal(common.FluidStorageClass))

			})
		})

		When("PVC already exists", func() {
			BeforeEach(func() {
				resources = append(resources, mockedObjects.PersistentVolumeClaim)
			})

			It("should not create PVC", func() {
				err := engine.createFusePersistentVolumeClaim()
				Expect(err).To(BeNil())
			})
		})

		When("Runtime sets extra labels and annotations for PVC", func() {
			BeforeEach(func() {
				metadata := datav1alpha1.Metadata{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels: map[string]string{
							"label1": "value1",
						},
						Annotations: map[string]string{
							"annotation1": "value2",
						},
					},
					Selector: metav1.GroupKind{
						Group: corev1.GroupName,
						Kind:  "PersistentVolumeClaim",
					},
				}
				newRuntimeInfo, err := base.BuildRuntimeInfo(alluxioruntime.Name, alluxioruntime.Namespace, common.AlluxioRuntime, base.WithMetadataList([]datav1alpha1.Metadata{metadata}))
				engine.runtimeInfo = newRuntimeInfo
				Expect(err).To(BeNil())
			})

			It("should create PVC with extra labels and annotations", func() {
				err := engine.createFusePersistentVolumeClaim()
				Expect(err).To(BeNil())

				gotPvc := &corev1.PersistentVolumeClaim{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotPvc)
				Expect(err).To(BeNil())
				Expect(gotPvc.Labels).Should(HaveKeyWithValue("label1", "value1"))
				Expect(gotPvc.Annotations).Should(HaveKeyWithValue("annotation1", "value2"))
			})
		})
	})
})
