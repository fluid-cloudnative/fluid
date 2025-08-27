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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine Volume Deletion Tests", Label("pkg.ddc.alluxio.delete_volume_test.go"), func() {
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
			mockedObjects.PersistentVolumeClaim,
			mockedObjects.PersistentVolume,
		}
	})

	// JustBeforeEach is guaranteed to run after every BeforeEach()
	// So it's easy to modify resources' specs with an extra BeforeEach()
	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = client
	})

	// TestAlluxioEngine_DeleteVolume tests the DeleteVolume function of the AlluxioEngine.
	// It sets up test cases with different PersistentVolume (PV) and PersistentVolumeClaim (PVC) inputs,
	// including scenarios with and without errors. The function uses a fake Kubernetes client to simulate
	// the behavior of the AlluxioEngine when deleting volumes. The test cases include:
	// 1. A common scenario where the volume should be deleted without errors.
	// 2. A scenario where an error is expected due to specific annotations on the PVC.
	// 3. A scenario where an error is expected because the AlluxioEngine is not running.
	// The function then runs these test cases using the doTestCases helper function to verify the expected outcomes.

	Describe("Test AlluxioEngine.DeleteVolume()", func() {
		BeforeEach(func() {
			mockedObjects.PersistentVolume.Namespace = ""
		})
		When("given AlluxioEngine works as expected", func() {
			It("should delete volume successfully", func() {
				err := engine.DeleteVolume()
				Expect(err).To(BeNil())

				err = client.Get(context.TODO(), types.NamespacedName{Namespace: engine.namespace, Name: engine.name}, &corev1.PersistentVolumeClaim{})
				Expect(apierrs.IsNotFound(err)).To(BeTrue())

				err = client.Get(context.TODO(), types.NamespacedName{Namespace: engine.namespace, Name: engine.runtimeInfo.GetPersistentVolumeName()}, &corev1.PersistentVolume{})
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})

		When("related PVC and PV is already deleted", func() {
			BeforeEach(func() {
				resources = []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
					// mockedObjects.PersistentVolumeClaim,
					// mockedObjects.PersistentVolume,
				}
			})

			It("don't need to do anything", func() {
				err := engine.DeleteVolume()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Test AlluxioEngine.deleteFusePersistentVolume()", func() {
		When("given AlluxioEngine works as expected", func() {
			BeforeEach(func() {
				mockedObjects.PersistentVolume.Namespace = ""
			})
			It("should delete fuse PV successfully", func() {
				err := engine.deleteFusePersistentVolume()
				Expect(err).To(BeNil())

				err = client.Get(context.TODO(), types.NamespacedName{Name: engine.runtimeInfo.GetPersistentVolumeName()}, &corev1.PersistentVolumeClaim{})
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})

		When("related PV is already deleted", func() {
			BeforeEach(func() {
				resources = []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
					mockedObjects.PersistentVolumeClaim,
					// mockedObjects.PersistentVolume,
				}

			})
			It("don't need to do anything", func() {
				err := engine.deleteFusePersistentVolume()
				Expect(err).To(BeNil())
			})
		})

		When("PV exists but does not have fluid annotations on it", func() {
			BeforeEach(func() {
				mockedObjects.PersistentVolume.Namespace = ""
				mockedObjects.PersistentVolume.Annotations = map[string]string{}
			})
			It("should not delete the PV", func() {
				err := engine.deleteFusePersistentVolume()
				Expect(err).To(BeNil())

				err = client.Get(context.TODO(), types.NamespacedName{Name: engine.runtimeInfo.GetPersistentVolumeName()}, &corev1.PersistentVolume{})
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Test AlluxioEngine.deleteFusePersistentVolumeClaim()", func() {
		When("given AlluxioEngine works as expected", func() {
			It("should delete the fuse PVC successfully", func() {
				err := engine.deleteFusePersistentVolumeClaim()
				Expect(err).To(BeNil())

				err = client.Get(context.TODO(), types.NamespacedName{Name: engine.name, Namespace: engine.namespace}, &corev1.PersistentVolumeClaim{})
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})

		When("related pvc is already deleted", func() {
			BeforeEach(func() {
				resources = []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
					// mockedObjects.PersistentVolumeClaim,
					mockedObjects.PersistentVolume,
				}
			})

			It("don't need to do anything", func() {
				err := engine.deleteFusePersistentVolumeClaim()
				Expect(err).To(BeNil())
			})
		})

		When("related pvc exists but do not have fluid annotaions on it", func() {
			BeforeEach(func() {
				mockedObjects.PersistentVolumeClaim.Annotations = map[string]string{}
			})

			It("should not delete the pvc", func() {
				err := engine.deleteFusePersistentVolumeClaim()
				Expect(err).To(BeNil())

				err = client.Get(context.TODO(), types.NamespacedName{Name: engine.name, Namespace: engine.namespace}, &corev1.PersistentVolumeClaim{})
				Expect(err).To(BeNil())
			})
		})
	})
})
