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
)

var _ = Describe("GooseFS Types", func() {
	Describe("getTiredStoreLevel0Path", func() {
		var (
			goosefs   *GooseFS
			name      string
			namespace string
		)

		BeforeEach(func() {
			name = "goosefs-01"
			namespace = "default"
		})

		Context("when tieredstore has level 0", func() {
			It("should return the configured path", func() {
				goosefs = &GooseFS{
					Tieredstore: Tieredstore{
						Levels: []Level{
							{
								Level: 0,
								Path:  "/mnt/demo/data",
							},
						},
					},
				}
				got := goosefs.getTiredStoreLevel0Path(name, namespace)
				Expect(got).To(Equal("/mnt/demo/data"))
			})
		})

		Context("when tieredstore has only level 1", func() {
			It("should return the default shm path", func() {
				goosefs = &GooseFS{
					Tieredstore: Tieredstore{
						Levels: []Level{
							{
								Level: 1,
								Path:  "/mnt/demo/data",
							},
						},
					},
				}
				got := goosefs.getTiredStoreLevel0Path(name, namespace)
				Expect(got).To(Equal("/dev/shm/default/goosefs-01"))
			})
		})

		Context("when tieredstore has multiple levels", func() {
			It("should return the level 0 path", func() {
				goosefs = &GooseFS{
					Tieredstore: Tieredstore{
						Levels: []Level{
							{
								Level: 1,
								Path:  "/mnt/ssd/data",
							},
							{
								Level: 0,
								Path:  "/mnt/mem/data",
							},
							{
								Level: 2,
								Path:  "/mnt/hdd/data",
							},
						},
					},
				}
				got := goosefs.getTiredStoreLevel0Path(name, namespace)
				Expect(got).To(Equal("/mnt/mem/data"))
			})
		})

		Context("when tieredstore is empty", func() {
			It("should return the default shm path", func() {
				goosefs = &GooseFS{
					Tieredstore: Tieredstore{
						Levels: []Level{},
					},
				}
				got := goosefs.getTiredStoreLevel0Path(name, namespace)
				Expect(got).To(Equal("/dev/shm/default/goosefs-01"))
			})
		})

		Context("with different namespace and name", func() {
			It("should construct the correct default path", func() {
				goosefs = &GooseFS{
					Tieredstore: Tieredstore{},
				}
				got := goosefs.getTiredStoreLevel0Path("mydata", "production")
				Expect(got).To(Equal("/dev/shm/production/mydata"))
			})
		})
	})
})
