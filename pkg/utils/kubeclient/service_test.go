/*

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

package kubeclient

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Use fake client because of it will be maintained in the long term
// due to https://github.com/kubernetes-sigs/controller-runtime/pull/1101
var _ = Describe("GetServiceByName", func() {
	var (
		namespace         string
		testServiceInputs []*v1.Service
		testServices      []runtime.Object
		mockClient        client.Client
	)

	BeforeEach(func() {
		namespace = "default"
		testServiceInputs = []*v1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "svc1"},
				Spec:       v1.ServiceSpec{},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "svc2", Annotations: common.GetExpectedFluidAnnotations()},
				Spec:       v1.ServiceSpec{},
			},
		}

		testServices = []runtime.Object{}
		for _, pv := range testServiceInputs {
			testServices = append(testServices, pv.DeepCopy())
		}

		mockClient = fake.NewFakeClientWithScheme(testScheme, testServices...)
	})

	Context("when service doesn't exist", func() {
		It("should not return an error", func() {
			_, err := GetServiceByName(mockClient, "notExist", namespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when service is not created by fluid", func() {
		It("should not return an error", func() {
			_, err := GetServiceByName(mockClient, "svc1", namespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
