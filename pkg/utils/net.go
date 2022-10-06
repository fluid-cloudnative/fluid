package utils

import (
	"bytes"
	"net"
	"sort"
)

func SortIpAddresses(ips []string) (orderedIps []string) {
	realIPs := make([]net.IP, 0, len(ips))

	for _, ip := range ips {
		realIPs = append(realIPs, net.ParseIP(ip))
	}

	sort.Slice(realIPs, func(i, j int) bool {
		return bytes.Compare(realIPs[i], realIPs[j]) < 0
	})

	for _, ip := range realIPs {
		orderedIps = append(orderedIps, ip.String())
	}

	return
}
