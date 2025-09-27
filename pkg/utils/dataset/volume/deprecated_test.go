/*
Copyright 2023 The Fluid Author.

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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
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

var _ = Describe("Deprecated PV Tests", Label("pkg.utils.dataset.volume.deprecated_test.go"), func() {
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
		_ = appsv1.AddToScheme(scheme)
		_ = datav1alpha1.AddToScheme(scheme)
		resources = nil
		log = fake.NullLogger()
	})

	JustBeforeEach(func() {
		clientObj = fake.NewFakeClientWithScheme(scheme, resources...)
	})

	When("deprecated PV exists", func() {
		BeforeEach(func() {
			resources = append(resources, &v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "hbase", Annotations: map[string]string{"CreatedBy": "fluid"}}})
			var err error
			runtimeInfo, err = base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
			Expect(err).To(BeNil())
		})
		It("should return true", func() {
			deprecated, err := HasDeprecatedPersistentVolumeName(clientObj, runtimeInfo, log)
			Expect(err).To(BeNil())
			Expect(deprecated).To(BeTrue())
		})
	})

	When("no deprecated PV exists", func() {
		BeforeEach(func() {
			var err error
			runtimeInfo, err = base.BuildRuntimeInfo("spark", "fluid", "alluxio")
			Expect(err).To(BeNil())
		})
		It("should return false", func() {
			deprecated, err := HasDeprecatedPersistentVolumeName(clientObj, runtimeInfo, log)
			Expect(err).To(BeNil())
			Expect(deprecated).To(BeFalse())
		})
	})
})
