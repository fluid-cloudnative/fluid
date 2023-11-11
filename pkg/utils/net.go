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
