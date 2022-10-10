package utils

import (
	"bytes"
	"net"
	"sort"
)

func SortIpAddresses(ips []string) (orderedIps []string) {
	realIPs := make([]net.IP, 0, len(ips))
	keys := make(map[string]bool)

	for _, ip := range ips {
		// Avoid duplicated keys
		if _, value := keys[ip]; !value {
			keys[ip] = true
			realIPs = append(realIPs, net.ParseIP(ip))
		}
	}

	sort.Slice(realIPs, func(i, j int) bool {
		return bytes.Compare(realIPs[i], realIPs[j]) < 0
	})

	for _, ip := range realIPs {
		orderedIps = append(orderedIps, ip.String())
	}

	return
}
