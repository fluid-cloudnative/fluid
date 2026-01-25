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
	"bytes"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testDatasetName      = "test-dataset"
	testDatasetNamespace = "test-namespace"
)

// testLogSink is a simple logger sink that writes to a buffer for testing
type testLogSink struct {
	buffer *bytes.Buffer
}

func (t *testLogSink) Init(info logr.RuntimeInfo) {}

func (t *testLogSink) Enabled(level int) bool {
	return true
}

func (t *testLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	t.buffer.WriteString(msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			t.buffer.WriteString(" ")
			t.buffer.WriteString(keysAndValues[i].(string))
			t.buffer.WriteString("=")
			// Simple string conversion
			switch v := keysAndValues[i+1].(type) {
			case string:
				t.buffer.WriteString(v)
			default:
				t.buffer.WriteString("%v")
			}
		}
	}
	t.buffer.WriteString("\n")
}

func (t *testLogSink) Error(err error, msg string, keysAndValues ...interface{}) {
	t.buffer.WriteString("ERROR: ")
	t.buffer.WriteString(msg)
	if err != nil {
		t.buffer.WriteString(" error=")
		t.buffer.WriteString(err.Error())
	}
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			t.buffer.WriteString(" ")
			t.buffer.WriteString(keysAndValues[i].(string))
			t.buffer.WriteString("=")
			// Simple string conversion
			switch v := keysAndValues[i+1].(type) {
			case string:
				t.buffer.WriteString(v)
			default:
				t.buffer.WriteString("%v")
			}
		}
	}
	t.buffer.WriteString("\n")
}

func (t *testLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return t
}

func (t *testLogSink) WithName(name string) logr.LogSink {
	return t
}

func newTestLogger() (logr.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	sink := &testLogSink{buffer: buf}
	return logr.New(sink), buf
}

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
		DescribeTable("should handle different input combinations",
			func(namespace, name, ufsTotal, fileNum string, shouldContain []string, shouldNotContain []string) {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					Done:      true,
					StartTime: time.Now(),
					UfsTotal:  ufsTotal,
					FileNum:   fileNum,
				}
				base.RecordDatasetMetrics(result, namespace, name, log)
				for _, s := range shouldContain {
					Expect(logBuf.String()).To(ContainSubstring(s))
				}
				for _, s := range shouldNotContain {
					Expect(logBuf.String()).NotTo(ContainSubstring(s))
				}
			},
			Entry("empty namespace",
				"", testDatasetName, "1GB", "100",
				[]string{"fail to validate RecordDatasetMetrics arguments", "datasetNamespace should not be empty"},
				[]string{},
			),
			Entry("empty name",
				testDatasetNamespace, "", "1GB", "100",
				[]string{"fail to validate RecordDatasetMetrics arguments", "datasetName should not be empty"},
				[]string{},
			),
			Entry("empty UfsTotal",
				testDatasetNamespace, testDatasetName, "", "100",
				[]string{},
				[]string{"ERROR"},
			),
			Entry("empty FileNum",
				testDatasetNamespace, testDatasetName, "1GB", "",
				[]string{},
				[]string{"ERROR"},
			),
			Entry("invalid UfsTotal format",
				testDatasetNamespace, testDatasetName, "invalid-size", "100",
				[]string{"fail to parse result.UfsTotal"},
				[]string{},
			),
			Entry("invalid FileNum format",
				testDatasetNamespace, testDatasetName, "1GB", "not-a-number",
				[]string{"fail to atoi result.FileNum"},
				[]string{},
			),
			Entry("all parameters valid",
				testDatasetNamespace, testDatasetName, "1GB", "100",
				[]string{},
				[]string{"ERROR"},
			),
			Entry("both UfsTotal and FileNum empty",
				testDatasetNamespace, testDatasetName, "", "",
				[]string{},
				[]string{"ERROR"},
			),
			Entry("both namespace and name empty",
				"", "", "1GB", "100",
				[]string{"fail to validate RecordDatasetMetrics arguments", "datasetNamespace should not be empty"},
				[]string{},
			),
		)
	})
})
