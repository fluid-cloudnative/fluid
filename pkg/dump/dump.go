package dump

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log logr.Logger

func init() {
	log = ctrl.Log.WithName("dump")
}

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

func InstallCoreDumpGenerator() {

	log.Info("Install goroutine generator")

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGQUIT)

	go func() {
		for {
			sig := <-signals

			switch sig {
			case syscall.SIGQUIT:
				t := time.Now()
				timestamp := fmt.Sprint(t.Format("20060102150405"))
				log.Info("User uses kill -3 to generate core dump")
				coredump("/tmp/go_" + timestamp + ".txt")
			// case syscall.SIGTERM:
			// 	fmt.Println("User told me to exit")
			// 	os.Exit(0)
			default:
				continue
			}
		}

	}()
}

func coredump(fileName string) {
	log.Info("Dump stacktrace to file", "fileName", fileName)
	trace := StackTrace(true)
	ioutil.WriteFile(fileName, []byte(trace), 0644)
	stdout := fmt.Sprintf("=== received SIGQUIT ===\n*** goroutine dump...\n%s", trace)
	log.Info(stdout)

}
