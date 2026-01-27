package discovery

import (
	"strconv"
	"strings"
	"sync"

	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	serverVersionMajor int
	serverVersionMinor int
	versionOnce        sync.Once
	discoveryErr       error
)

// GetServerVersion returns the major and minor version of the kubernetes server
func GetServerVersion() (int, int, error) {
	versionOnce.Do(func() {
		restConfig, err := ctrl.GetConfig()
		if err != nil {
			discoveryErr = err
			return
		}
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
		if err != nil {
			discoveryErr = err
			return
		}
		versionInfo, err := discoveryClient.ServerVersion()
		if err != nil {
			discoveryErr = err
			return
		}

		serverVersionMajor, err = strconv.Atoi(versionInfo.Major)
		if err != nil {
			discoveryErr = err
			return
		}

		// Minor version might have a suffix like "28+", trim it
		minor := strings.TrimSuffix(versionInfo.Minor, "+")
		serverVersionMinor, err = strconv.Atoi(minor)
		if err != nil {
			discoveryErr = err
			return
		}
	})

	return serverVersionMajor, serverVersionMinor, discoveryErr
}
