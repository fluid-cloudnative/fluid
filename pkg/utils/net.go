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
