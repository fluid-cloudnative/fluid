package utils

import (
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
	"time"
)

func TestNewPprofServer(t *testing.T) {
	setupLog := ctrl.Log.WithName("test")
	pprofAddr := "127.0.0.1:6060"
	NewPprofServer(setupLog, pprofAddr)
	paths := []string{
		"/debug/pprof/", "/debug/pprof/cmdline", "/debug/pprof/profile", "/debug/pprof/symbol", "/debug/pprof/trace",
	}
	time.Sleep(10 * time.Second)
	for _, p := range paths {
		resp, err := http.Get("http://127.0.0.1:6060" + p)
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status code is %d", resp.StatusCode)
		}
	}
}
