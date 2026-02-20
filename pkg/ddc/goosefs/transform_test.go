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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	fakeutils "github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("TransformFuse", func() {
	BeforeEach(func() {
		ctrl.SetLogger(zap.New(func(o *zap.Options) {
			o.Development = true
		}))
	})

	type testCase struct {
		runtime *datav1alpha1.GooseFSRuntime
		dataset *datav1alpha1.Dataset
		value   *GooseFS
		expect  []string
	}

	DescribeTable("should transform fuse configuration correctly",
		func(tc testCase) {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "goosefs")
			Expect(err).NotTo(HaveOccurred())

			engine := &GooseFSEngine{
				runtimeInfo: runtimeInfo,
				Client:      fakeutils.NewFakeClientWithScheme(testScheme),
				Log:         ctrl.Log,
			}

			err = engine.transformFuse(tc.runtime, tc.dataset, tc.value)
			Expect(err).NotTo(HaveOccurred())
			Expect(tc.value.Fuse.Args).To(Equal(tc.expect))
		},
		Entry("with owner UID and GID",
			func() testCase {
				var x int64 = 1000
				return testCase{
					runtime: &datav1alpha1.GooseFSRuntime{
						Spec: datav1alpha1.GooseFSRuntimeSpec{
							Fuse: datav1alpha1.GooseFSFuseSpec{},
						},
					},
					dataset: &datav1alpha1.Dataset{
						Spec: datav1alpha1.DatasetSpec{
							Mounts: []datav1alpha1.Mount{{
								MountPoint: "local:///mnt/test",
								Name:       "test",
							}},
							Owner: &datav1alpha1.User{
								UID: &x,
								GID: &x,
							},
						},
					},
					value:  &GooseFS{},
					expect: []string{"fuse", "--fuse-opts=rw,direct_io,uid=1000,gid=1000,allow_other"},
				}
			}(),
		),
	)
})
