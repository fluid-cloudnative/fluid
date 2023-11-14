/*
Copyright 2021 The Fluid Authors.

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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AllocatePolicy string

const (
	Random AllocatePolicy = "random"
	BitMap AllocatePolicy = "bitmap"
)

func ValidateEnum(allocatePolicyStr string) (AllocatePolicy, error) {
	switch AllocatePolicy(allocatePolicyStr) {
	case Random, BitMap:
		return AllocatePolicy(allocatePolicyStr), nil
	default:
		return AllocatePolicy(allocatePolicyStr), fmt.Errorf("runtime-port-allocator can only be random or bitmap")
	}
}

type BatchAllocatorInterface interface {
	Allocate(int) error

	Release(int) error

	AllocateBatch(portNum int) ([]int, error)
	needResetReservedPorts() bool
}

// RuntimePortAllocator is an allocator resonsible for maintaining port usage information
// given a user-defined port range. It allocates and releases ports when a port is requested or
// reclaimed by a runtime.
type RuntimePortAllocator struct {
	pa             BatchAllocatorInterface
	allocatePolicy AllocatePolicy

	client client.Client
	pr     *net.PortRange

	// getReservedPorts is a func helps the RuntimePortAllocator restore port usage information
	// from the cluster status. The func depends on specific implementation for different Runtimes.
	getReservedPorts func(client client.Client) (ports []int, err error)

	log logr.Logger
}

// rpa is a global singleton of type RuntimePortAllocator
var rpa *RuntimePortAllocator

// SetupRuntimePortAllocator instantiates the global singleton rpa, set up port allocating policy according to the given allocatePolicyStr.
// Currently the valid policies are either "random" or "bitmap".
func SetupRuntimePortAllocator(client client.Client, pr *net.PortRange, allocatePolicyStr string, getReservedPorts func(client client.Client) (ports []int, err error)) error {
	policy, err := ValidateEnum(allocatePolicyStr)
	if err != nil {
		return err
	}
	SetupRuntimePortAllocatorWithType(client, pr, policy, getReservedPorts)
	return nil
}

// SetupRuntimePortAllocatorWithType instantiates the global singleton rpa with specified port allocating policy
func SetupRuntimePortAllocatorWithType(client client.Client, pr *net.PortRange, allocatePolicy AllocatePolicy, getReservedPorts func(client client.Client) (ports []int, err error)) {
	rpa = &RuntimePortAllocator{client: client, pr: pr, allocatePolicy: allocatePolicy, getReservedPorts: getReservedPorts}
	rpa.log = ctrl.Log.WithName("RuntimePortAllocator")
}

// GetRuntimePortAllocator restore the port allocator and gets the global singleton. This should be the only way others access
// the RuntimePortAllocator and it must be called after SetupRuntimePortAllocator
func GetRuntimePortAllocator() (*RuntimePortAllocator, error) {
	if rpa.pa == nil {
		if err := rpa.createAndRestorePortAllocator(); err != nil {
			return nil, err
		}
	}
	return rpa, nil
}

// createAndRestorePortAllocator creates and restores port allocator with runtime-specific logic
func (alloc *RuntimePortAllocator) createAndRestorePortAllocator() (err error) {
	switch alloc.allocatePolicy {
	case Random:
		alloc.pa, err = newRandomAllocator(alloc.pr, alloc.log)
	case BitMap:
		alloc.pa, err = newBitMapAllocator(alloc.pr, alloc.log)
	default:
		err = errors.New("runtime-port-allocator can only be random or bitmap")
	}

	if err != nil {
		return err
	}

	// policy should check reserved ports
	if alloc.pa.needResetReservedPorts() {
		ports, err := alloc.getReservedPorts(alloc.client)
		if err != nil {
			return err
		}
		alloc.log.Info("Found reserved ports", "ports", ports)

		for _, port := range ports {
			if err = alloc.pa.Allocate(port); err != nil {
				alloc.log.Error(err, "can't allocate reserved ports", "port", port)
			}
		}
	}

	return nil
}

// GetAvailablePorts requests portNum ports from the port allocator.
// It returns an int array with allocated ports in it.
func (alloc *RuntimePortAllocator) GetAvailablePorts(portNum int) (ports []int, err error) {
	if alloc.pa == nil {
		return nil, errors.New("Runtime port allocator not setup")
	}

	ports, err = alloc.pa.AllocateBatch(portNum)
	if err != nil {
		return ports, err
	}

	alloc.log.Info("Successfully allocated ports", "expeceted port num", portNum, "allocated ports", ports)
	return ports, nil
}

// ReleaseReservedPorts releases all the ports in the given int array.
func (alloc *RuntimePortAllocator) ReleaseReservedPorts(ports []int) {
	alloc.log.Info("Releasing reserved ports", "ports to be released", ports)
	for _, port := range ports {
		if err := alloc.pa.Release(port); err != nil {
			alloc.log.Error(err, "can't release port", "port", port)
		}
	}
}
