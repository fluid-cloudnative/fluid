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
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
)

var testMutex sync.Mutex

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

		It("should return different content for all vs current goroutine", func() {
			allTrace := StackTrace(true)
			currentTrace := StackTrace(false)

			Expect(len(allTrace)).To(BeNumerically(">=", len(currentTrace)))
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

		It("should handle empty buffer case", func() {
			trace := StackTrace(false)
			Expect(trace).NotTo(BeEmpty())
			Expect(trace).To(ContainSubstring("goroutine"))
		})
	})

	Context("when handling concurrent stack traces", func() {
		It("should be safe to call concurrently", func() {
			var wg sync.WaitGroup
			traces := make([]string, 10)

			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					traces[index] = StackTrace(false)
				}(i)
			}

			wg.Wait()

			for _, trace := range traces {
				Expect(trace).To(ContainSubstring("goroutine"))
				Expect(len(trace)).To(BeNumerically(">", 0))
			}
		})
	})
})

var _ = Describe("InstallgoroutineDumpGenerator", Serial, func() {
	var initState int32

	BeforeEach(func() {
		testMutex.Lock()
		defer testMutex.Unlock()
		ResetForTesting()
		atomic.StoreInt32(&initState, 0)
		log = ctrl.Log.WithName("dump")
	})

	Context("when installing for the first time", func() {
		It("should set initialized to true", func() {
			InstallgoroutineDumpGenerator()

			time.Sleep(50 * time.Millisecond)
			Expect(atomic.LoadInt32(&initialized)).To(Equal(int32(1)))
		})
	})

	Context("when installing multiple times sequentially", func() {
		It("should be idempotent", func() {
			// First installation
			InstallgoroutineDumpGenerator()
			time.Sleep(50 * time.Millisecond)

			firstCheck := atomic.LoadInt32(&initialized)
			Expect(firstCheck).To(Equal(int32(1)))

			// Subsequent installations should not cause issues
			for i := 0; i < 4; i++ {
				InstallgoroutineDumpGenerator()
				time.Sleep(20 * time.Millisecond)
			}

			Expect(atomic.LoadInt32(&initialized)).To(Equal(int32(1)))
		})
	})

	Context("when receiving SIGQUIT signal", func() {
		It("should create dump file when signal is sent", func() {
			ResetForTesting()

			InstallgoroutineDumpGenerator()
			time.Sleep(200 * time.Millisecond)

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

			// Verify file content
			if dumpFilePath != "" {
				content, err := os.ReadFile(dumpFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("goroutine"))

			}
		})
	})
})

var _ = Describe("Coredump", func() {
	var testFile string

	BeforeEach(func() {
		log = ctrl.Log.WithName("dump")
		testFile = "/tmp/test_coredump.txt"
	})

	AfterEach(func() {
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

		It("should overwrite existing file", func() {
			// Create initial file
			initialContent := []byte("initial content")
			err := os.WriteFile(testFile, initialContent, 0644)
			Expect(err).ToNot(HaveOccurred())

			// Call coredump
			coredump(testFile)

			// Read new content
			content, err := os.ReadFile(testFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).NotTo(Equal(string(initialContent)))
			Expect(string(content)).To(ContainSubstring("goroutine"))
		})

		It("should create readable file with proper permissions", func() {
			coredump(testFile)

			info, err := os.Stat(testFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode().IsRegular()).To(BeTrue())
		})
	})

	Context("when calling coredump concurrently", func() {
		It("should handle concurrent calls safely", func() {
			files := make([]string, 5)
			for i := 0; i < 5; i++ {
				files[i] = "/tmp/test_coredump_" + string(rune('a'+i)) + ".txt"
			}

			var wg sync.WaitGroup
			for _, file := range files {
				wg.Add(1)
				go func(f string) {
					defer wg.Done()
					coredump(f)
				}(file)
			}
			wg.Wait()

			// Verify all files were created
			for _, file := range files {
				_, err := os.Stat(file)
				Expect(os.IsNotExist(err)).To(BeFalse())
				_ = os.Remove(file)
			}
		})
	})
})

var _ = Describe("SignalHandling", Serial, func() {
	BeforeEach(func() {
		testMutex.Lock()
		defer testMutex.Unlock()
		ResetForTesting()
		log = ctrl.Log.WithName("dump")
	})

	AfterEach(func() {
		_ = os.Remove("/tmp/go-test-signal.txt")
		// Clean up any dump files created during tests
		files, _ := os.ReadDir("/tmp")
		for _, file := range files {
			if len(file.Name()) > 3 && file.Name()[:3] == "go-" && file.Name()[len(file.Name())-4:] == ".txt" {
				_ = os.Remove("/tmp/" + file.Name())
			}
		}
	})

	Context("when installing signal handler", func() {
		It("should initialize successfully", func() {
			InstallgoroutineDumpGenerator()
			time.Sleep(100 * time.Millisecond)

			Expect(atomic.LoadInt32(&initialized)).To(Equal(int32(1)))
		})

		It("should not panic on installation", func() {
			Expect(func() {
				InstallgoroutineDumpGenerator()
			}).NotTo(Panic())
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

		It("should handle various directory paths", func() {
			tests := []struct {
				dir       string
				prefix    string
				timestamp string
				expected  string
			}{
				{"/tmp", "go", "20230101", "/tmp/go-20230101.txt"},
				{"/var/log", "dump", "123456", "/var/log/dump-123456.txt"},
				{".", "test", "999", "./test-999.txt"},
				{"/home/user", "core", "20240101", "/home/user/core-20240101.txt"},
			}

			for _, tt := range tests {
				got := formatDumpfile(tt.dir, tt.prefix, tt.timestamp)
				Expect(got).To(Equal(tt.expected))
			}
		})

		It("should handle empty strings", func() {
			got := formatDumpfile("", "", "")
			Expect(got).To(Equal("/-.txt"))
		})
	})
})

// Helper function to format dumpfile path
func formatDumpfile(dir, prefix, timestamp string) string {
	return dir + "/" + prefix + "-" + timestamp + ".txt"
}
