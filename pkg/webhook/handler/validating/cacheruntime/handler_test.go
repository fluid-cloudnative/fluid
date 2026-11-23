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

package cacheruntime

import (
	"context"
	"encoding/json"
	"net/http"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("SpecUpdateHandler", func() {
	var (
		handler    *SpecUpdateHandler
		oldRuntime *datav1alpha1.CacheRuntime
		newRuntime *datav1alpha1.CacheRuntime
	)

	raw := func(obj runtime.Object) runtime.RawExtension {
		data, err := json.Marshal(obj)
		Expect(err).NotTo(HaveOccurred())
		return runtime.RawExtension{Raw: data}
	}

	request := func(op admissionv1.Operation, oldObj, newObj runtime.Object) admission.Request {
		req := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: op,
			Namespace: "default",
			Name:      "demo",
		}}
		if oldObj != nil {
			req.OldObject = raw(oldObj)
		}
		if newObj != nil {
			req.Object = raw(newObj)
		}
		return req
	}

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(scheme)).To(Succeed())
		handler = &SpecUpdateHandler{}
		handler.Setup(nil, nil, admission.NewDecoder(scheme))

		spec := baseSpec()
		oldRuntime = &datav1alpha1.CacheRuntime{
			TypeMeta:   metav1.TypeMeta{Kind: datav1alpha1.CacheRuntimeKind, APIVersion: datav1alpha1.GroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "default"},
			Spec:       spec,
		}
		newRuntime = oldRuntime.DeepCopy()
	})

	It("allows an update that only touches updatable fields", func() {
		newRuntime.Spec.Worker.Replicas = 3
		newRuntime.Spec.Worker.RuntimeVersion.ImageTag = "v2"

		resp := handler.Handle(context.TODO(), request(admissionv1.Update, oldRuntime, newRuntime))
		Expect(resp.Allowed).To(BeTrue())
	})

	It("allows the controller's own writes to metadata and status", func() {
		newRuntime.Finalizers = []string{"fluid.io/cacheruntime"}
		newRuntime.Labels = map[string]string{"fluid.io/dataset": "demo"}
		newRuntime.Status.Worker.Phase = datav1alpha1.RuntimePhaseReady

		resp := handler.Handle(context.TODO(), request(admissionv1.Update, oldRuntime, newRuntime))
		Expect(resp.Allowed).To(BeTrue())
	})

	It("denies an update that changes an immutable field and names the field", func() {
		newRuntime.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota.Set(4 << 30)

		resp := handler.Handle(context.TODO(), request(admissionv1.Update, oldRuntime, newRuntime))
		Expect(resp.Allowed).To(BeFalse())
		Expect(resp.Result.Code).To(Equal(int32(http.StatusForbidden)))
		Expect(resp.Result.Message).To(ContainSubstring(UpdatableFieldsHint))
		Expect(resp.Result.Message).To(ContainSubstring("spec.worker.tieredStore"))
		Expect(resp.Result.Message).NotTo(ContainSubstring("spec.worker.replicas"))
	})

	It("denies an update that mixes updatable and immutable fields", func() {
		newRuntime.Spec.Worker.Replicas = 3
		newRuntime.Spec.Client.Options["debug"] = "true"

		resp := handler.Handle(context.TODO(), request(admissionv1.Update, oldRuntime, newRuntime))
		Expect(resp.Allowed).To(BeFalse())
		Expect(resp.Result.Message).To(ContainSubstring("spec.client.options"))
	})

	It("lists every immutable field that changed", func() {
		newRuntime.Spec.RuntimeClassName = "other"
		newRuntime.Spec.Master.Args = nil

		resp := handler.Handle(context.TODO(), request(admissionv1.Update, oldRuntime, newRuntime))
		Expect(resp.Allowed).To(BeFalse())
		Expect(resp.Result.Message).To(ContainSubstring("spec.runtimeClassName"))
		Expect(resp.Result.Message).To(ContainSubstring("spec.master.args"))
	})

	DescribeTable("allows operations other than UPDATE without decoding",
		func(op admissionv1.Operation) {
			resp := handler.Handle(context.TODO(), request(op, nil, nil))
			Expect(resp.Allowed).To(BeTrue())
		},
		Entry("CREATE", admissionv1.Create),
		Entry("DELETE", admissionv1.Delete),
		Entry("CONNECT", admissionv1.Connect),
	)

	It("allows a subresource write without comparing the spec", func() {
		req := request(admissionv1.Update, oldRuntime, newRuntime)
		req.SubResource = "scale"
		// a Scale object, not a CacheRuntime, would be carried here
		req.Object = runtime.RawExtension{Raw: []byte(`{"spec":{"replicas":3}}`)}

		resp := handler.Handle(context.TODO(), req)
		Expect(resp.Allowed).To(BeTrue())
	})

	DescribeTable("rejects a request whose objects cannot be decoded",
		func(corrupt func(req *admission.Request)) {
			req := request(admissionv1.Update, oldRuntime, newRuntime)
			corrupt(&req)

			resp := handler.Handle(context.TODO(), req)
			Expect(resp.Allowed).To(BeFalse())
			Expect(resp.Result.Code).To(Equal(int32(http.StatusBadRequest)))
		},
		Entry("missing old object", func(req *admission.Request) { req.OldObject = runtime.RawExtension{} }),
		Entry("missing new object", func(req *admission.Request) { req.Object = runtime.RawExtension{} }),
		Entry("malformed new object", func(req *admission.Request) { req.Object = runtime.RawExtension{Raw: []byte(`{`)} }),
	)

	It("uses the validator's field errors verbatim", func() {
		newRuntime.Spec.Client.CleanPolicy = datav1alpha1.OnDemandCleanPolicy
		want := ValidateSpecUpdate(&oldRuntime.Spec, &newRuntime.Spec)
		Expect(want).To(HaveLen(1))
		Expect(want[0].Type).To(Equal(field.ErrorTypeForbidden))

		resp := handler.Handle(context.TODO(), request(admissionv1.Update, oldRuntime, newRuntime))
		Expect(resp.Result.Message).To(ContainSubstring(want.ToAggregate().Error()))
	})
})
