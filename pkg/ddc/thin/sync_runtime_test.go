/*
  Copyright 2022 The Fluid Authors.

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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/pkg/runtime"
)

var _ = Describe("ThinEngine_SyncRuntime", func() {
	It("should return false for changed and no error", func() {
		var ctx runtime.ReconcileRequestContext
		engine := ThinEngine{}

		gotChanged, err := engine.SyncRuntime(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})
})
