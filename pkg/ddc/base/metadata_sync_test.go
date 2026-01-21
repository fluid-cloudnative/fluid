/*
Copyright 2023 The Fluid Authors.

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

package base_test

import (
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testDatasetName      = "test-dataset"
	testDatasetNamespace = "test-namespace"
)

var _ = Describe("MetadataSync", func() {
	Describe("SafeClose", func() {
		Context("when channel is nil", func() {
			It("should return false without panic", func() {
				var nilCh chan base.MetadataSyncResult
				closed := base.SafeClose(nilCh)
				Expect(closed).To(BeFalse())
			})
		})

		Context("when channel is open", func() {
			It("should close the channel and return false", func() {
				openCh := make(chan base.MetadataSyncResult)
				closed := base.SafeClose(openCh)
				Expect(closed).To(BeFalse())
			})
		})

		Context("when channel is already closed", func() {
			It("should return true without panic", func() {
				closedCh := make(chan base.MetadataSyncResult)
				close(closedCh)
				closed := base.SafeClose(closedCh)
				Expect(closed).To(BeTrue())
			})
		})
	})

	Describe("SafeSend", func() {
		Context("when channel is nil", func() {
			It("should return false without panic", func() {
				var nilCh chan base.MetadataSyncResult
				closed := base.SafeSend(nilCh, base.MetadataSyncResult{})
				Expect(closed).To(BeFalse())
			})
		})

		Context("when channel is open", func() {
			It("should send to the channel and return false", func() {
				openCh := make(chan base.MetadataSyncResult, 1) // Buffer 1 to prevent blocking
				result := base.MetadataSyncResult{
					Done:      true,
					StartTime: time.Now(),
					UfsTotal:  "1GB",
					FileNum:   "100",
				}
				closed := base.SafeSend(openCh, result)
				Expect(closed).To(BeFalse())
				received := <-openCh
				Expect(received.Done).To(BeTrue())
				Expect(received.UfsTotal).To(Equal("1GB"))
				Expect(received.FileNum).To(Equal("100"))
			})
		})

		Context("when channel is already closed", func() {
			It("should return true without panic", func() {
				closedCh := make(chan base.MetadataSyncResult)
				close(closedCh)
				closed := base.SafeSend(closedCh, base.MetadataSyncResult{})
				Expect(closed).To(BeTrue())
			})
		})
	})

	Describe("RecordDatasetMetrics", func() {
		var log = fake.NullLogger()

		Context("when datasetNamespace is empty", func() {
			It("should return early without panic", func() {
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "100",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, "", testDatasetName, log)
				}).NotTo(Panic())
			})
		})

		Context("when datasetName is empty", func() {
			It("should return early without panic", func() {
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "100",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, testDatasetNamespace, "", log)
				}).NotTo(Panic())
			})
		})

		Context("when UfsTotal is empty", func() {
			It("should skip parsing UfsTotal without panic", func() {
				result := base.MetadataSyncResult{
					UfsTotal: "",
					FileNum:  "100",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				}).NotTo(Panic())
			})
		})

		Context("when FileNum is empty", func() {
			It("should skip parsing FileNum without panic", func() {
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				}).NotTo(Panic())
			})
		})

		Context("when UfsTotal has invalid format", func() {
			It("should log error without panic", func() {
				result := base.MetadataSyncResult{
					UfsTotal: "invalid-size",
					FileNum:  "100",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				}).NotTo(Panic())
			})
		})

		Context("when FileNum has invalid format", func() {
			It("should log error without panic", func() {
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "not-a-number",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				}).NotTo(Panic())
			})
		})

		Context("when all parameters are valid", func() {
			It("should record metrics without panic", func() {
				result := base.MetadataSyncResult{
					Done:      true,
					StartTime: time.Now(),
					UfsTotal:  "1GB",
					FileNum:   "100",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				}).NotTo(Panic())
			})
		})

		Context("when both UfsTotal and FileNum are empty", func() {
			It("should complete without panic", func() {
				result := base.MetadataSyncResult{
					Done:      true,
					StartTime: time.Now(),
					UfsTotal:  "",
					FileNum:   "",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				}).NotTo(Panic())
			})
		})

		Context("when both namespace and name are empty", func() {
			It("should return early without panic", func() {
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "100",
				}
				Expect(func() {
					base.RecordDatasetMetrics(result, "", "", log)
				}).NotTo(Panic())
			})
		})
	})
})
