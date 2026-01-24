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
		Context("when datasetNamespace is empty", func() {
			It("should log error about invalid datasetNamespace", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "100",
				}
				base.RecordDatasetMetrics(result, "", testDatasetName, log)
				Expect(logBuf.String()).To(ContainSubstring("fail to validate RecordDatasetMetrics arguments"))
				Expect(logBuf.String()).To(ContainSubstring("datasetNamespace should not be empty"))
			})
		})

		Context("when datasetName is empty", func() {
			It("should log error about invalid datasetName", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "100",
				}
				base.RecordDatasetMetrics(result, testDatasetNamespace, "", log)
				Expect(logBuf.String()).To(ContainSubstring("fail to validate RecordDatasetMetrics arguments"))
				Expect(logBuf.String()).To(ContainSubstring("datasetName should not be empty"))
			})
		})

		Context("when UfsTotal is empty", func() {
			It("should skip parsing UfsTotal without error", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					UfsTotal: "",
					FileNum:  "100",
				}
				base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				Expect(logBuf.String()).NotTo(ContainSubstring("ERROR"))
			})
		})

		Context("when FileNum is empty", func() {
			It("should skip parsing FileNum without error", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "",
				}
				base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				Expect(logBuf.String()).NotTo(ContainSubstring("ERROR"))
			})
		})

		Context("when UfsTotal has invalid format", func() {
			It("should log error about parsing failure", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					UfsTotal: "invalid-size",
					FileNum:  "100",
				}
				base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				Expect(logBuf.String()).To(ContainSubstring("fail to parse result.UfsTotal"))
			})
		})

		Context("when FileNum has invalid format", func() {
			It("should log error about atoi failure", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "not-a-number",
				}
				base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				Expect(logBuf.String()).To(ContainSubstring("fail to atoi result.FileNum"))
			})
		})

		Context("when all parameters are valid", func() {
			It("should record metrics without error", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					Done:      true,
					StartTime: time.Now(),
					UfsTotal:  "1GB",
					FileNum:   "100",
				}
				base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				Expect(logBuf.String()).NotTo(ContainSubstring("ERROR"))
			})
		})

		Context("when both UfsTotal and FileNum are empty", func() {
			It("should complete without error", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					Done:      true,
					StartTime: time.Now(),
					UfsTotal:  "",
					FileNum:   "",
				}
				base.RecordDatasetMetrics(result, testDatasetNamespace, testDatasetName, log)
				Expect(logBuf.String()).NotTo(ContainSubstring("ERROR"))
			})
		})

		Context("when both namespace and name are empty", func() {
			It("should log error about invalid datasetNamespace", func() {
				log, logBuf := newTestLogger()
				result := base.MetadataSyncResult{
					UfsTotal: "1GB",
					FileNum:  "100",
				}
				base.RecordDatasetMetrics(result, "", "", log)
				Expect(logBuf.String()).To(ContainSubstring("fail to validate RecordDatasetMetrics arguments"))
				Expect(logBuf.String()).To(ContainSubstring("datasetNamespace should not be empty"))
			})
		})
	})
})
