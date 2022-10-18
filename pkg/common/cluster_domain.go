package common

import (
	"errors"
	"os"
	"strings"
)

// GetClusterDomain get cluster domain: cluster.local from /etc/resolv.conf
func GetClusterDomain() (string, error) {
	resolveConf, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return "", err
	}
	return parseResolvConf(string(resolveConf))
}

// parseResolvConf parse cluster domain from /etc/resolv.conf
// search default.svc.cluster.local svc.cluster.local cluster.local
// for how k8s generate `resolv.conf`, ref:
// https://github.com/kubernetes/kubernetes/blob/542ec977054c16c7981606cb1590cc39154ddf01/pkg/kubelet/network/dns/dns.go#L167
func parseResolvConf(conf string) (string, error) {
	for _, line := range strings.Split(conf, "\n") {
		line := strings.TrimSpace(line)
		if strings.HasPrefix(line, "search") {
			search := strings.Split(line, " ")
			if len(search) >= 4 {
				return search[3], nil
			}
		}
	}
	return "", errors.New("failed to parse cluster domain from resolv.conf")
}
