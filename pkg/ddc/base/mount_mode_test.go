/*
Copyright 2024 The Fluid Authors.

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

package base

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseMountModeSelectorFromStr", func() {
	DescribeTable("parsing mount mode strings",
		func(input string, wantModes []MountMode, wantErr bool) {
			got, err := ParseMountModeSelectorFromStr(input)
			if wantErr {
				Expect(err).To(HaveOccurred())
				return
			}
			Expect(err).NotTo(HaveOccurred())
			for _, mode := range wantModes {
				Expect(got.Selected(mode)).To(BeTrue(), "missing mode %v", mode)
			}
			Expect(len(got)).To(Equal(len(wantModes)))
		},
		Entry("empty string returns empty selector", "", []MountMode{}, false),
		Entry("All selects all modes", "All", SupportedMountModes, false),
		Entry("None returns empty selector", "None", []MountMode{}, false),
		Entry("MountPod selects only MountPod", "MountPod", []MountMode{MountPodMountMode}, false),
		Entry("Sidecar selects only Sidecar", "Sidecar", []MountMode{SidecarMountMode}, false),
		Entry("comma separated selects multiple", "MountPod,Sidecar", []MountMode{MountPodMountMode, SidecarMountMode}, false),
		Entry("unsupported mode returns error", "InvalidMode", nil, true),
		Entry("All with other modes takes precedence", "MountPod,All", SupportedMountModes, false),
		Entry("None with other modes takes precedence", "None,Sidecar", []MountMode{}, false),
		Entry("None terminates selection", "MountPod,None", []MountMode{MountPodMountMode}, false),
		Entry("whitespace in modes returns error", "MountPod, Sidecar", nil, true),
		Entry("partial invalid mode returns error", "MountPod,InvalidMode", nil, true),
	)
})

var _ = Describe("mountModeSelector.Selected", func() {
	It("should return true for existing mode", func() {
		selector := mountModeSelector{MountPodMountMode: true}
		Expect(selector.Selected(MountPodMountMode)).To(BeTrue())
	})

	It("should return false for non-existing mode", func() {
		selector := mountModeSelector{MountPodMountMode: true}
		Expect(selector.Selected(SidecarMountMode)).To(BeFalse())
	})

	It("should return false for empty selector", func() {
		emptySelector := mountModeSelector{}
		Expect(emptySelector.Selected(MountPodMountMode)).To(BeFalse())
	})
})
