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

package goosefs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

var _ = Describe("GooseFSEngine_SyncRuntime", func() {
	type testCase struct {
		name        string
		ctx         cruntime.ReconcileRequestContext
		wantChanged bool
		wantErr     bool
	}

	DescribeTable("should sync runtime correctly",
		func(tc testCase) {
			e := &GooseFSEngine{}
			gotChanged, err := e.SyncRuntime(tc.ctx)

			if tc.wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(gotChanged).To(Equal(tc.wantChanged))
		},
		Entry("default case",
			testCase{
				name:        "default",
				wantChanged: false,
				wantErr:     false,
			},
		),
	)
})
