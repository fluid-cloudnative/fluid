/*
Copyright 2022 The Fluid Author.

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

package deploy

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindofsx"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/thin"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/vineyard"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const controllerNamespace = common.NamespaceFluidSystem

var _ = Describe("runtime controller scaleout", func() {
	var originalPodNamespace string
	var hadOriginalPodNamespace bool
	var originalResolveDefaultPrecheckFuncs func() map[string]CheckFunc

	BeforeEach(func() {
		originalPodNamespace, hadOriginalPodNamespace = os.LookupEnv(common.MyPodNamespace)
		originalResolveDefaultPrecheckFuncs = resolveDefaultPrecheckFuncs
	})

	AfterEach(func() {
		setPrecheckFunc(nil)
		resolveDefaultPrecheckFuncs = originalResolveDefaultPrecheckFuncs
		restoreEnv(common.MyPodNamespace, originalPodNamespace, hadOriginalPodNamespace)
	})

	Describe("scaleoutDeploymentIfNeeded", func() {
		It("returns an error when the controller deployment is missing", func() {
			fakeClient := newFakeClient()

			scaled, err := scaleoutDeploymentIfNeeded(fakeClient, types.NamespacedName{
				Namespace: corev1.NamespaceDefault,
				Name:      "missing-controller",
			}, fake.NullLogger())

			Expect(err).To(HaveOccurred())
			Expect(scaled).To(BeFalse())
		})

		DescribeTable("scales deployments according to replica rules",
			func(deployment *appsv1.Deployment, wantScaled bool, wantReplicas int32) {
				fakeClient := newFakeClient(deployment)

				scaled, err := scaleoutDeploymentIfNeeded(fakeClient, types.NamespacedName{
					Namespace: deployment.Namespace,
					Name:      deployment.Name,
				}, fake.NullLogger())

				Expect(err).NotTo(HaveOccurred())
				Expect(scaled).To(Equal(wantScaled))

				stored := &appsv1.Deployment{}
				Expect(fakeClient.Get(context.TODO(), types.NamespacedName{
					Namespace: deployment.Namespace,
					Name:      deployment.Name,
				}, stored)).To(Succeed())
				Expect(*stored.Spec.Replicas).To(Equal(wantReplicas))
			},
			Entry("defaults zero replicas to one when no annotation exists",
				newDeployment("unknown-controller", 0, nil), true, int32(1)),
			Entry("uses the configured replica annotation when it is greater than one",
				newDeployment("goosefsruntime-controller", 0, map[string]string{common.RuntimeControllerReplicas: "3"}), true, int32(3)),
			Entry("enforces a minimum of one replica when annotation is zero",
				newDeployment("juicefsruntime-controller", 0, map[string]string{common.RuntimeControllerReplicas: "0"}), true, int32(1)),
			Entry("leaves already running controllers unchanged",
				newDeployment("jindoruntime-controller", 1, nil), false, int32(1)),
		)
	})

	Describe("ScaleoutRuntimeControllerOnDemand", func() {
		BeforeEach(func() {
			Expect(os.Setenv(common.MyPodNamespace, controllerNamespace)).To(Succeed())
		})

		It("returns no matched controller when no runtime precheck matches the dataset", func() {
			fakeClient := newFakeClient(runtimeObjects()...)
			setPrecheckFunc(runtimePrecheckFuncs())

			controllerName, scaled, err := ScaleoutRuntimeControllerOnDemand(fakeClient, types.NamespacedName{
				Namespace: corev1.NamespaceDefault,
				Name:      "notFound",
			}, fake.NullLogger())

			Expect(err).To(MatchError("no matched controller for dataset default/notFound"))
			Expect(controllerName).To(BeEmpty())
			Expect(scaled).To(BeFalse())
		})

		It("returns the matched controller name and scales the deployment", func() {
			fakeClient := newFakeClient(append(runtimeObjects(), controllerDeployments()...)...)
			setPrecheckFunc(runtimePrecheckFuncs())

			controllerName, scaled, err := ScaleoutRuntimeControllerOnDemand(fakeClient, types.NamespacedName{
				Namespace: corev1.NamespaceDefault,
				Name:      "alluxio",
			}, fake.NullLogger())

			Expect(err).NotTo(HaveOccurred())
			Expect(controllerName).To(Equal("alluxioruntime-controller"))
			Expect(scaled).To(BeTrue())

			stored := &appsv1.Deployment{}
			Expect(fakeClient.Get(context.TODO(), types.NamespacedName{
				Namespace: controllerNamespace,
				Name:      controllerName,
			}, stored)).To(Succeed())
			Expect(*stored.Spec.Replicas).To(Equal(int32(1)))
		})

		It("leaves already running matched controllers unchanged", func() {
			fakeClient := newFakeClient(append(runtimeObjects(), controllerDeployments()...)...)
			setPrecheckFunc(runtimePrecheckFuncs())

			controllerName, scaled, err := ScaleoutRuntimeControllerOnDemand(fakeClient, types.NamespacedName{
				Namespace: corev1.NamespaceDefault,
				Name:      "jindo",
			}, fake.NullLogger())

			Expect(err).NotTo(HaveOccurred())
			Expect(controllerName).To(Equal("jindoruntime-controller"))
			Expect(scaled).To(BeFalse())

			stored := &appsv1.Deployment{}
			Expect(fakeClient.Get(context.TODO(), types.NamespacedName{
				Namespace: controllerNamespace,
				Name:      controllerName,
			}, stored)).To(Succeed())
			Expect(*stored.Spec.Replicas).To(Equal(int32(1)))
		})

		It("keeps package-global precheck functions isolated between tests", func() {
			fakeClient := newFakeClient(controllerDeployments()...)
			setPrecheckFunc(map[string]CheckFunc{
				"custom-controller": func(client.Client, types.NamespacedName) (bool, error) {
					return true, nil
				},
			})

			controllerName, scaled, err := ScaleoutRuntimeControllerOnDemand(fakeClient, types.NamespacedName{
				Namespace: corev1.NamespaceDefault,
				Name:      "ignored",
			}, fake.NullLogger())

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("custom-controller")))
			Expect(controllerName).To(BeEmpty())
			Expect(scaled).To(BeFalse())
		})
	})

	Describe("precheck function isolation", func() {
		It("does not retain caller mutations after setting precheck funcs", func() {
			checks := runtimePrecheckFuncs()
			setPrecheckFunc(checks)

			delete(checks, "alluxioruntime-controller")

			Expect(getPrecheckFuncs()).To(HaveKey("alluxioruntime-controller"))
		})

		It("returns a snapshot that callers can mutate without changing stored precheck funcs", func() {
			setPrecheckFunc(runtimePrecheckFuncs())

			checks := getPrecheckFuncs()
			delete(checks, "alluxioruntime-controller")

			Expect(getPrecheckFuncs()).To(HaveKey("alluxioruntime-controller"))
		})

		It("does not pin discovery-filtered defaults into package-global state", func() {
			resolveDefaultPrecheckFuncs = func() map[string]CheckFunc {
				return runtimePrecheckFuncs()
			}
			setPrecheckFunc(nil)

			checks := getPrecheckFuncs()

			Expect(checks).NotTo(BeNil())
			Expect(precheckFuncs).To(BeNil())
		})
	})
})

func newFakeClient(objects ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	Expect(appsv1.AddToScheme(scheme)).To(Succeed())
	Expect(datav1alpha1.AddToScheme(scheme)).To(Succeed())
	return fake.NewFakeClientWithScheme(scheme, objects...)
}

func newDeployment(name string, replicas int32, annotations map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   controllerNamespace,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
		},
	}
}

func controllerDeployments() []runtime.Object {
	return []runtime.Object{
		newDeployment("alluxioruntime-controller", 0, nil),
		newDeployment("jindoruntime-controller", 1, nil),
		newDeployment("juicefsruntime-controller", 0, map[string]string{common.RuntimeControllerReplicas: "0"}),
		newDeployment("goosefsruntime-controller", 0, map[string]string{common.RuntimeControllerReplicas: "3"}),
		newDeployment("unknown-controller", 0, nil),
	}
}

func runtimeObjects() []runtime.Object {
	return []runtime.Object{
		&datav1alpha1.AlluxioRuntime{ObjectMeta: metav1.ObjectMeta{Name: "alluxio", Namespace: corev1.NamespaceDefault}},
		&datav1alpha1.GooseFSRuntime{ObjectMeta: metav1.ObjectMeta{Name: "goosefs", Namespace: corev1.NamespaceDefault}},
		&datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{Name: "jindo", Namespace: corev1.NamespaceDefault}},
		&datav1alpha1.JuiceFSRuntime{ObjectMeta: metav1.ObjectMeta{Name: "juicefs", Namespace: corev1.NamespaceDefault}},
	}
}

func runtimePrecheckFuncs() map[string]CheckFunc {
	return map[string]CheckFunc{
		"alluxioruntime-controller":  alluxio.Precheck,
		"jindoruntime-controller":    jindofsx.Precheck,
		"juicefsruntime-controller":  juicefs.Precheck,
		"goosefsruntime-controller":  goosefs.Precheck,
		"thinruntime-controller":     thin.Precheck,
		"efcruntime-controller":      efc.Precheck,
		"vineyardruntime-controller": vineyard.Precheck,
	}
}

func restoreEnv(key, value string, hadValue bool) {
	if hadValue {
		Expect(os.Setenv(key, value)).To(Succeed())
		return
	}

	Expect(os.Unsetenv(key)).To(Succeed())
}
