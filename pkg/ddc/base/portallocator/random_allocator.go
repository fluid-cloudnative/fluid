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

package portallocator

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/net"
)

type RandomAllocator struct {
	portRange *net.PortRange
	// lock make rand thread safe
	lock sync.Mutex
	rand *rand.Rand
	log  logr.Logger
}

func (r *RandomAllocator) needResetReservedPorts() bool {
	return false
}

func newRandomAllocator(pr *net.PortRange, log logr.Logger) (*RandomAllocator, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &RandomAllocator{
		portRange: pr,
		rand:      r,
		log:       log,
	}, nil
}

func (r *RandomAllocator) Allocate(port int) error {
	// not judge whether port can be allocated or not
	return nil

}

func (r *RandomAllocator) Release(i int) error {
	// no need to release
	return nil
}

func (r *RandomAllocator) AllocateBatch(portNum int) (ports []int, err error) {
	var availPort int
	var allocatedPorts = map[int]bool{}

	// prevent infinite for loop
	if portNum > r.portRange.Size {
		return ports, errors.New("required port size exceeds the configured size")
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for i := 0; i < portNum; {
		availPort = r.portRange.Base + r.rand.Intn(r.portRange.Size)
		_, ok := allocatedPorts[availPort]
		if !ok {
			i++
			allocatedPorts[availPort] = true
			ports = append(ports, availPort)
		}
	}

	return
}
