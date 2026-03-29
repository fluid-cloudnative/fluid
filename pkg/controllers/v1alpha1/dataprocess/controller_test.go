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

package dataprocess

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// newTestDataProcessReconciler builds a DataProcessReconciler for unit tests.
func newTestDataProcessReconciler(s *runtime.Scheme, objs ...runtime.Object) *DataProcessReconciler {
	if s == nil {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	}
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	log := logf.Log.WithName("dataprocess-test")
	recorder := record.NewFakeRecorder(32)
	return NewDataProcessReconciler(fakeClient, log, s, recorder)
}

var _ = Describe("DataProcessReconciler", func() {

	Describe("ControllerName", func() {
		It("should return the expected controller name", func() {
			r := newTestDataProcessReconciler(nil)
			Expect(r.ControllerName()).To(Equal("DataProcessReconciler"))
		})
	})

	Describe("NewDataProcessReconciler", func() {
		It("should create a non-nil reconciler with a Scheme", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			r := newTestDataProcessReconciler(s)
			Expect(r).NotTo(BeNil())
			Expect(r.Scheme).NotTo(BeNil())
		})
	})

	Describe("Build", func() {
		It("should return a dataProcessOperation for a valid DataProcess object", func() {
			r := newTestDataProcessReconciler(nil)
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op, err := r.Build(dp)
			Expect(err).NotTo(HaveOccurred())
			Expect(op).NotTo(BeNil())
		})

		It("should return an error when given a non-DataProcess object", func() {
			r := newTestDataProcessReconciler(nil)
			ds := &datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{Name: "ds", Namespace: "default"},
			}
			op, err := r.Build(ds)
			Expect(err).To(HaveOccurred())
			Expect(op).To(BeNil())
		})
	})

	Describe("Reconcile", func() {
		It("should return no error when the DataProcess is not found", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			r := newTestDataProcessReconciler(s)
			// No DataProcess objects registered — should get NotFound and return cleanly.
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "missing", Namespace: "default"},
			}
			result, err := r.Reconcile(context.TODO(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("should reconcile successfully when DataProcess exists", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: datav1alpha1.DataProcessSpec{
					Dataset: datav1alpha1.TargetDatasetWithMountPath{
						TargetDataset: datav1alpha1.TargetDataset{
							Name:      "ds",
							Namespace: "default",
						},
						MountPath: "/data",
					},
					Processor: datav1alpha1.Processor{
						Script: &datav1alpha1.ScriptProcessor{
							Source: "echo hello",
						},
					},
				},
			}
			r := newTestDataProcessReconciler(s, dp)
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
			}
			result, err := r.Reconcile(context.TODO(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{RequeueAfter: 20 * time.Second}))
		})
	})
})
