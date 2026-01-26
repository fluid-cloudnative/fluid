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
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("dump") 
var initialized int32 // Changed to atomic int32

var dumpfileMutex sync.RWMutex
var dumpfile string = "%s-%s.txt"

func StackTrace(all bool) string {
	buf := make([]byte, 10240)

	for {
		size := runtime.Stack(buf, all)

		if size == len(buf) {
			buf = make([]byte, len(buf)<<1)
			continue
		}
		break

	}

	return string(buf)
}

func InstallgoroutineDumpGenerator() {
	if !atomic.CompareAndSwapInt32(&initialized, 0, 1) {
		log.Info("Do nothing for installing grouting dump.")
		return
	}

	log.Info("Register goroutine dump generator")

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGQUIT)

	go func() {
		for {
			sig := <-signals

			switch sig {
			case syscall.SIGQUIT:
				t := time.Now()
				timestamp := fmt.Sprint(t.Format("20060102150405"))
				log.Info("User uses kill -3 to generate goroutine dump")
				dumpfileMutex.RLock()
				filename := fmt.Sprintf(dumpfile, "/tmp", "go", timestamp)
				dumpfileMutex.RUnlock()
				coredump(filename)
			default:
				continue
			}
		}

	}()
}

func coredump(fileName string) {
	log.Info("Dump stacktrace to file", "fileName", fileName)
	trace := StackTrace(true)
	err := os.WriteFile(fileName, []byte(trace), 0644)
	if err != nil {
		log.Error(err, "Failed to write coredump.")
	}
	stdout := fmt.Sprintf("=== received SIGQUIT ===\n*** goroutine dump...\n%s", trace)
	log.Info(stdout)

}

// ResetForTesting resets the package state for testing purposes
// This should only be used in tests
func ResetForTesting() {
	atomic.StoreInt32(&initialized, 0)
}
