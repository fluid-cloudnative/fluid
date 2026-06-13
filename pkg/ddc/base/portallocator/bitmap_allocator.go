/*
Copyright 2023 The Fluid Author.

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

package portallocator

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/kubernetes/pkg/registry/core/service/allocator"
	"k8s.io/kubernetes/pkg/registry/core/service/portallocator"
)

type BitMapAllocator struct {
	alloc *portallocator.PortAllocator
	log   logr.Logger
}

func (b *BitMapAllocator) needResetReservedPorts() bool {
	return true
}

// newBitMapAllocator creates a bitmap-based port allocator for the given port range.
// It initializes the underlying allocator and returns a BitMapAllocator instance.
// It returns an error if the allocator cannot be created.
func newBitMapAllocator(pr *net.PortRange, log logr.Logger) (BatchAllocatorInterface, error) {
	// TODO passing offset value.
	alloc, err := portallocator.New(*pr, func(max int, rangeSpec string, offset int) (allocator.Interface, error) {
		return allocator.NewAllocationMap(max, rangeSpec), nil
	})

	if err != nil {
		return nil, err
	}

	return &BitMapAllocator{
		alloc: alloc,
		log:   log,
	}, nil
}

func (b *BitMapAllocator) Allocate(port int) error {
	return b.alloc.Allocate(port)
}

func (b *BitMapAllocator) Release(port int) error {
	return b.alloc.Release(port)
}

// AllocateBatch allocates a batch of ports from the bitmap allocator.
// This function attempts to allocate the specified number of ports and returns them as a slice.
// If allocation fails for any reason (e.g., insufficient available ports), it will rollback all
// previously allocated ports in this batch and exit the process with status code 1 to ensure
// port allocation state consistency.
//
// Parameters:
//   - portNum (int): The number of ports to allocate.
//
// Returns:
//   - ports ([]int): A slice containing the allocated port numbers.
//   - err (error): Always nil if the function returns (note: function exits process on failure)
func (b *BitMapAllocator) AllocateBatch(portNum int) (ports []int, err error) {
	var availPort int

	for i := 0; i < portNum; i++ {
		if availPort, err = b.alloc.AllocateNext(); err != nil {
			b.log.Error(err, "can't allocate next, all ports are in use")
			break
		} else {
			ports = append(ports, availPort)
		}
	}
	// Something unexpected happened, rollback to release allocated ports
	if err != nil || len(ports) < portNum {
		for _, reservedPort := range ports {
			_ = b.Release(reservedPort)
		}
		// Allocated port may not be released as expect, restart to restore allocated ports.
		b.log.Error(errors.Errorf("can't get enough available ports, only %d ports are available", len(ports)), "")
		b.log.Info("Exit to restore port allocator...")
		os.Exit(1)
	}

	return
}
