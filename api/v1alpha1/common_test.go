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

package v1alpha1

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("common API helpers", func() {
	Describe("MetadataSyncPolicy.AutoSyncEnabled", func() {
		It("defaults to true when auto sync is not configured", func() {
			policy := &MetadataSyncPolicy{}

			Expect(policy.AutoSyncEnabled()).To(BeTrue())
		})

		It("returns true when auto sync is explicitly enabled", func() {
			enabled := true
			policy := &MetadataSyncPolicy{AutoSync: &enabled}

			Expect(policy.AutoSyncEnabled()).To(BeTrue())
		})

		It("returns false when auto sync is explicitly disabled", func() {
			disabled := false
			policy := &MetadataSyncPolicy{AutoSync: &disabled}

			Expect(policy.AutoSyncEnabled()).To(BeFalse())
		})
	})

	Describe("Dataset.CanbeBound", func() {
		It("returns true when no runtime is recorded", func() {
			dataset := &Dataset{}

			Expect(dataset.CanbeBound("runtime", "fluid", common.AccelerateCategory)).To(BeTrue())
		})

		It("returns true only for a matching runtime identity", func() {
			dataset := &Dataset{
				Status: DatasetStatus{
					Runtimes: []Runtime{
						{Name: "target", Namespace: "fluid", Category: common.AccelerateCategory},
					},
				},
			}

			Expect(dataset.CanbeBound("target", "fluid", common.AccelerateCategory)).To(BeTrue())
			Expect(dataset.CanbeBound("other", "fluid", common.AccelerateCategory)).To(BeFalse())
		})
	})

	Describe("Dataset.IsExclusiveMode", func() {
		It("treats default placement as exclusive", func() {
			dataset := &Dataset{}

			Expect(dataset.IsExclusiveMode()).To(BeTrue())
		})

		It("returns false for shared placement mode", func() {
			dataset := &Dataset{Spec: DatasetSpec{PlacementMode: ShareMode}}

			Expect(dataset.IsExclusiveMode()).To(BeFalse())
		})
	})
})
