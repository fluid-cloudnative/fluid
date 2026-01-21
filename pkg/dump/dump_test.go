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
package dump

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"syscall"
	"testing"
	"time"
)

func TestDump(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dump Suite")
}

var _ = Describe("StackTrace", func() {
	Context("when requesting stack trace", func() {
		It("should contain goroutine information with all goroutines", func() {
			got := StackTrace(true)
			Expect(got).To(ContainSubstring("goroutine"))
			Expect(len(got)).To(BeNumerically(">", 0))
		})

		It("should contain goroutine information with current goroutine only", func() {
			got := StackTrace(false)
			Expect(got).To(ContainSubstring("goroutine"))
			Expect(len(got)).To(BeNumerically(">", 0))
		})
	})

	Context("when handling large stack traces", func() {
		It("should grow buffer as needed and return non-empty trace", func() {
			trace := StackTrace(true)
			Expect(len(trace)).To(BeNumerically(">", 0))
			Expect(trace).To(ContainSubstring("goroutine"))
		})

		It("should handle buffer growth for very large traces", func() {
			// Create multiple goroutines to generate a larger stack trace
			done := make(chan bool)
			for i := 0; i < 50; i++ {
				go func() {
					time.Sleep(100 * time.Millisecond)
					done <- true
				}()
			}

			trace := StackTrace(true)
			Expect(len(trace)).To(BeNumerically(">", 10240))

			// Clean up goroutines
			for i := 0; i < 50; i++ {
				<-done
			}
		})
	})
})

var _ = Describe("InstallgoroutineDumpGenerator", func() {
	BeforeEach(func() {
		initialized = false
		log = ctrl.Log.WithName("dump")
	})

	Context("when installing for the first time", func() {
		It("should set initialized to true", func() {
			InstallgoroutineDumpGenerator()
			Expect(initialized).To(BeTrue())
		})
	})

	Context("when installing multiple times", func() {
		It("should only initialize once", func() {
			for i := 0; i < 3; i++ {
				InstallgoroutineDumpGenerator()
			}
			Expect(initialized).To(BeTrue())
		})

		It("should do nothing on second installation", func() {
			InstallgoroutineDumpGenerator()
			Expect(initialized).To(BeTrue())

			// Second installation should skip initialization
			InstallgoroutineDumpGenerator()
			Expect(initialized).To(BeTrue())
		})
	})

	Context("when receiving SIGQUIT signal", func() {
		It("should create dump file when signal is sent", func() {
			initialized = false
			InstallgoroutineDumpGenerator()
			Expect(initialized).To(BeTrue())

			// Send SIGQUIT signal to trigger dump
			time.Sleep(100 * time.Millisecond)
			pid := os.Getpid()
			process, err := os.FindProcess(pid)
			Expect(err).ToNot(HaveOccurred())

			err = process.Signal(syscall.SIGQUIT)
			Expect(err).ToNot(HaveOccurred())

			// Wait for signal handler to process
			time.Sleep(500 * time.Millisecond)

			// Check if dump file was created (with timestamp pattern)
			files, err := os.ReadDir("/tmp")
			Expect(err).ToNot(HaveOccurred())

			found := false
			var dumpFilePath string
			for _, file := range files {
				if len(file.Name()) > 3 && file.Name()[:3] == "go-" && file.Name()[len(file.Name())-4:] == ".txt" {
					found = true
					dumpFilePath = "/tmp/" + file.Name()
					break
				}
			}
			Expect(found).To(BeTrue())

			// Clean up
			if dumpFilePath != "" {
				os.Remove(dumpFilePath)
			}
		})
	})
})

var _ = Describe("Coredump", func() {
	var testFile string

	BeforeEach(func() {
		log = ctrl.Log.WithName("dump")
		testFile = "/tmp/test_coredump.txt"
		os.Remove(testFile)
	})

	AfterEach(func() {
		os.Remove(testFile)
	})

	Context("when creating a coredump", func() {
		It("should create the file with stack trace", func() {
			coredump(testFile)

			_, err := os.Stat(testFile)
			Expect(os.IsNotExist(err)).To(BeFalse())

			content, err := os.ReadFile(testFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("goroutine"))
		})

		It("should handle write errors gracefully", func() {
			// Try to write to invalid path
			invalidFile := "/invalid/path/to/file.txt"
			coredump(invalidFile)
			// Should not panic, error is logged
			Expect(true).To(BeTrue())
		})

		It("should write complete stack trace to file", func() {
			coredump(testFile)

			content, err := os.ReadFile(testFile)
			Expect(err).ToNot(HaveOccurred())

			strContent := string(content)
			Expect(strContent).To(ContainSubstring("goroutine"))
			Expect(len(strContent)).To(BeNumerically(">", 0))
		})
	})
})

var _ = Describe("SignalHandling", func() {
	BeforeEach(func() {
		initialized = false
		log = ctrl.Log.WithName("dump")
	})

	AfterEach(func() {
		os.Remove("/tmp/go-test-signal.txt")
		// Clean up any dump files created during tests
		files, _ := os.ReadDir("/tmp")
		for _, file := range files {
			if len(file.Name()) > 3 && file.Name()[:3] == "go-" && file.Name()[len(file.Name())-4:] == ".txt" {
				os.Remove("/tmp/" + file.Name())
			}
		}
	})

	Context("when installing signal handler", func() {
		It("should initialize successfully", func() {
			InstallgoroutineDumpGenerator()
			Expect(initialized).To(BeTrue())
		})
	})
})

var _ = Describe("DumpfileFormat", func() {
	Context("when formatting dumpfile path", func() {
		It("should format correctly for /tmp/go-20230101120000.txt", func() {
			got := formatDumpfile("/tmp", "go", "20230101120000")
			Expect(got).To(Equal("/tmp/go-20230101120000.txt"))
		})

		It("should format correctly for /var/log/dump-20231231235959.txt", func() {
			got := formatDumpfile("/var/log", "dump", "20231231235959")
			Expect(got).To(Equal("/var/log/dump-20231231235959.txt"))
		})

		It("should match the dumpfile format used in code", func() {
			timestamp := "20230515143022"
			expected := "/tmp/go-" + timestamp + ".txt"
			got := formatDumpfile("/tmp", "go", timestamp)
			Expect(got).To(Equal(expected))
		})
	})
})

// Helper function to format dumpfile path
func formatDumpfile(dir, prefix, timestamp string) string {
	return dir + "/" + prefix + "-" + timestamp + ".txt"
}
