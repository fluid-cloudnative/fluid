package discovery

import (
	"log"
	"strconv"
	"strings"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

func SupportsNativeSidecarOrDefault(cfg *rest.Config, defaultValue bool) bool {
	if cfg == nil {
		return defaultValue
	}
	return SupportsNativeSidecar(cfg)
}

func SupportsNativeSidecar(cfg *rest.Config) bool {
	// Add nil check
	if cfg == nil {
		return false
	}
	
	// fmt.Printf("SupportsNativeSidecar: cfg.Host = %q\n", cfg.Host)
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		log.Printf("Failed to create discovery client for native sidecar detection: %v", err)
		return false
	}

	info, err := dc.ServerVersion()
	if err != nil {
		return false
	}

	majorStr := strings.TrimPrefix(info.Major, "v")
	minorStr := strings.TrimSuffix(info.Minor, "+")

	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return false
	}

	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return false
	}

	// native sidecar supported from k8s 1.29+
	if major > 1 {
		return true
	}

	return major == 1 && minor >= 29
}