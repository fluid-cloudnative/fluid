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

package mutating

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandlerMap", func() {
	It("should register exactly one handler under WebhookSchedulePodPath", func() {
		Expect(HandlerMap).To(HaveLen(1))
		Expect(HandlerMap).To(HaveKey(common.WebhookSchedulePodPath))
	})

	It("should map WebhookSchedulePodPath to a *FluidMutatingHandler", func() {
		handler, ok := HandlerMap[common.WebhookSchedulePodPath]
		Expect(ok).To(BeTrue())
		Expect(handler).To(BeAssignableToTypeOf(&FluidMutatingHandler{}))
	})
})
