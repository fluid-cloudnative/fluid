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

package thin

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ThinEngine bindDatasetToMountedPersistentVolumeClaim", func() {
	newEngine := func(objs ...runtime.Object) *ThinEngine {
		return &ThinEngine{
			name:      "dataset",
			namespace: "fluid",
			Client:    fake.NewFakeClientWithScheme(testScheme, objs...),
			Log:       fake.NullLogger(),
		}
	}

	newDataset := func(mounts ...datav1alpha1.Mount) *datav1alpha1.Dataset {
		return &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dataset",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{Mounts: mounts},
		}
	}

	newPVC := func(name string) *corev1.PersistentVolumeClaim {
		return &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "PersistentVolumeClaim",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "fluid",
				UID:       types.UID(name + "-uid"),
			},
		}
	}

	It("adds the mounted pvc as a dataset owner reference", func() {
		pvc := newPVC("mounted-pvc")
		engine := newEngine(
			newDataset(datav1alpha1.Mount{Name: "native", MountPoint: "pvc://mounted-pvc"}),
			pvc,
		)

		Expect(engine.bindDatasetToMountedPersistentVolumeClaim()).To(Succeed())

		updatedDataset := &datav1alpha1.Dataset{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "dataset", Namespace: "fluid"}, updatedDataset)).To(Succeed())
		Expect(updatedDataset.OwnerReferences).To(ContainElement(metav1.OwnerReference{
			APIVersion: pvc.APIVersion,
			Kind:       pvc.Kind,
			Name:       pvc.Name,
			UID:        pvc.UID,
		}))
	})

	It("does not duplicate an existing pvc owner reference", func() {
		pvc := newPVC("mounted-pvc")
		ownerReference := metav1.OwnerReference{
			APIVersion: pvc.APIVersion,
			Kind:       pvc.Kind,
			Name:       pvc.Name,
			UID:        pvc.UID,
		}
		dataset := newDataset(datav1alpha1.Mount{Name: "native", MountPoint: "pvc://mounted-pvc"})
		dataset.OwnerReferences = []metav1.OwnerReference{ownerReference}

		engine := newEngine(dataset, pvc)

		Expect(engine.bindDatasetToMountedPersistentVolumeClaim()).To(Succeed())

		updatedDataset := &datav1alpha1.Dataset{}
		Expect(engine.Get(context.TODO(), types.NamespacedName{Name: "dataset", Namespace: "fluid"}, updatedDataset)).To(Succeed())
		Expect(updatedDataset.OwnerReferences).To(HaveLen(1))
		Expect(updatedDataset.OwnerReferences[0]).To(Equal(ownerReference))
	})

	It("returns an error when more than one pvc mount is declared", func() {
		engine := newEngine(
			newDataset(
				datav1alpha1.Mount{Name: "first", MountPoint: "pvc://mounted-pvc-1"},
				datav1alpha1.Mount{Name: "second", MountPoint: "pvc://mounted-pvc-2"},
			),
			newPVC("mounted-pvc-1"),
			newPVC("mounted-pvc-2"),
		)

		err := engine.bindDatasetToMountedPersistentVolumeClaim()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("can only contain one pvc:// mount point"))
	})
})
