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

package datamigrate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("DataMigrateReconciler", func() {
	Describe("NewDataMigrateReconciler", func() {
		It("should create a reconciler with the provided fields", func() {
			scheme := runtime.NewScheme()
			fakeClient := fake.NewFakeClientWithScheme(scheme)
			log := fake.NullLogger()
			recorder := record.NewFakeRecorder(10)

			r := NewDataMigrateReconciler(fakeClient, log, scheme, recorder)

			Expect(r).NotTo(BeNil())
			Expect(r.Scheme).To(Equal(scheme))
			Expect(r.OperationReconciler).NotTo(BeNil())
		})
	})

	Describe("ControllerName", func() {
		It("should return DataMigrateReconciler", func() {
			scheme := runtime.NewScheme()
			fakeClient := fake.NewFakeClientWithScheme(scheme)
			log := fake.NullLogger()
			recorder := record.NewFakeRecorder(10)

			r := NewDataMigrateReconciler(fakeClient, log, scheme, recorder)

			Expect(r.ControllerName()).To(Equal("DataMigrateReconciler"))
		})
	})

	Describe("Build", func() {
		It("should return a dataMigrateOperation when given a DataMigrate object", func() {
			scheme := runtime.NewScheme()
			fakeClient := fake.NewFakeClientWithScheme(scheme)
			log := fake.NullLogger()
			recorder := record.NewFakeRecorder(10)

			r := NewDataMigrateReconciler(fakeClient, log, scheme, recorder)

			dm := &datav1alpha1.DataMigrate{}
			dm.Name = "test-migrate"
			dm.Namespace = "default"

			op, err := r.Build(dm)

			Expect(err).NotTo(HaveOccurred())
			Expect(op).NotTo(BeNil())
		})

		It("should return an error when given a non-DataMigrate object", func() {
			scheme := runtime.NewScheme()
			fakeClient := fake.NewFakeClientWithScheme(scheme)
			log := fake.NullLogger()
			recorder := record.NewFakeRecorder(10)

			r := NewDataMigrateReconciler(fakeClient, log, scheme, recorder)

			notDM := &datav1alpha1.DataLoad{}

			op, err := r.Build(notDM)

			Expect(err).To(HaveOccurred())
			Expect(op).To(BeNil())
		})
	})
})
