/*
Copyright 2023 The Fluid Authors.

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

package volume

import (
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete Volume Tests", Label("pkg.utils.dataset.volume.delete_test.go"), func() {
	var (
		scheme      *runtime.Scheme
		clientObj   client.Client
		runtimeInfo base.RuntimeInfoInterface
		resources   []runtime.Object
		log         logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = v1.AddToScheme(scheme)
		resources = nil
		log = fake.NullLogger()
	})

	JustBeforeEach(func() {
		clientObj = fake.NewFakeClientWithScheme(scheme, resources...)
	})

	Context("Test DeleteFusePersistentVolume()", func() {
		When("PV is not annotated by fluid", func() {
			BeforeEach(func() {
				resources = append(resources, &v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "no-anno"}})
				var err error
				runtimeInfo, err = base.BuildRuntimeInfo("no-anno", "fluid", "alluxio")
				Expect(err).To(BeNil())
			})
			It("should no-op and return success", func() {
				Expect(DeleteFusePersistentVolume(clientObj, runtimeInfo, log)).To(Succeed())
			})
		})

		When("PV with fluid annotations exists", func() {
			BeforeEach(func() {
				resources = append(resources, &v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "fluid-hadoop", Annotations: common.GetExpectedFluidAnnotations()}})
				var err error
				runtimeInfo, err = base.BuildRuntimeInfo("hadoop", "fluid", "alluxio")
				Expect(err).To(BeNil())
			})
			It("should delete the PV successfully", func() {
				Expect(DeleteFusePersistentVolume(clientObj, runtimeInfo, log)).To(Succeed())
			})
		})
	})

	Context("Test DeleteFusePersistentVolumeClaim()", func() {
		When("PVC is stuck terminating with pvc-protection finalizer", func() {
			BeforeEach(func() {
				pvc := &v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "force-delete", Namespace: "fluid", Finalizers: []string{"kubernetes.io/pvc-protection"}, Annotations: common.GetExpectedFluidAnnotations(), DeletionTimestamp: &metav1.Time{Time: time.Now().Add(-35 * time.Second)}}}
				resources = append(resources, pvc)
				var err error
				runtimeInfo, err = base.BuildRuntimeInfo("force-delete", "fluid", "alluxio")
				Expect(err).To(BeNil())
			})
			It("should remove finalizer if needed and succeed", func() {
				Expect(DeleteFusePersistentVolumeClaim(clientObj, runtimeInfo, log)).To(Succeed())
			})
		})
	})
})
