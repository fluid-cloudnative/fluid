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

package databackup

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("DataBackupReconciler", func() {

	Describe("ControllerName", func() {
		It("should return the constant controller name", func() {
			r := &DataBackupReconciler{}
			Expect(r.ControllerName()).To(Equal(controllerName))
		})
	})

	Describe("NewDataBackupReconciler", func() {
		It("should initialize reconciler with all required fields set", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			fakeClient := fake.NewFakeClientWithScheme(s)
			log := ctrl.Log.WithName("test")
			recorder := record.NewFakeRecorder(10)

			r := NewDataBackupReconciler(fakeClient, log, s, recorder)
			Expect(r).NotTo(BeNil())
			Expect(r.Scheme).To(Equal(s))
			Expect(r.OperationReconciler).NotTo(BeNil())
		})
	})

	Describe("Build", func() {
		It("should return a dataBackupOperation for a valid DataBackup object", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			fakeClient := fake.NewFakeClientWithScheme(s)
			log := ctrl.Log.WithName("test")
			recorder := record.NewFakeRecorder(10)
			r := NewDataBackupReconciler(fakeClient, log, s, recorder)

			dataBackup := &datav1alpha1.DataBackup{}
			op, err := r.Build(dataBackup)
			Expect(err).NotTo(HaveOccurred())
			Expect(op).NotTo(BeNil())
		})

		It("should return an error for a non-DataBackup object", func() {
			s := runtime.NewScheme()
			fakeClient := fake.NewFakeClientWithScheme(s)
			log := ctrl.Log.WithName("test")
			recorder := record.NewFakeRecorder(10)
			r := NewDataBackupReconciler(fakeClient, log, s, recorder)

			dataset := &datav1alpha1.Dataset{}
			op, err := r.Build(dataset)
			Expect(err).To(HaveOccurred())
			Expect(op).To(BeNil())
		})
	})
})
