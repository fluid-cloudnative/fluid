/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
