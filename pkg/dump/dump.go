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
	"syscall"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log logr.Logger

var initialized bool

var dumpfile string = "%s/%s-%s.txt"

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
	if !initialized {
		log = ctrl.Log.WithName("dump")
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
					// coredump("/tmp/go_" + timestamp + ".txt")
					coredump(fmt.Sprintf(dumpfile, "/tmp", "go", timestamp))
				// case syscall.SIGTERM:
				// 	fmt.Println("User told me to exit")
				// 	os.Exit(0)
				default:
					continue
				}
			}

		}()

		initialized = true
	} else {
		log.Info("Do nothing for installing grouting dump.")
	}
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
