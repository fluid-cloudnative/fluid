package portallocator

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/kubernetes/pkg/registry/core/service/allocator"
	"k8s.io/kubernetes/pkg/registry/core/service/portallocator"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RuntimePortAllocator is an allocator resonsible for maintaining port usage information
// given a user-defined port range. It allocates and releases ports when a port is requested or
// reclaimed by a runtime.
type RuntimePortAllocator struct {
	pa     *portallocator.PortAllocator
	client client.Client
	pr     *net.PortRange

	// getReservedPorts is a func helps the RuntimePortAllocator restore port usage information
	// from the cluster status. The func depends on specific implementation for different Runtimes.
	getReservedPorts func(client client.Client) (ports []int, err error)

	log logr.Logger
}

// rpa is a global singleton of type RuntimePortAllocator
var rpa *RuntimePortAllocator

// SetupRuntimePortAllocator instantiates the global singleton rpa
func SetupRuntimePortAllocator(client client.Client, pr *net.PortRange, getReservedPorts func(client client.Client) (ports []int, err error)) {
	rpa = &RuntimePortAllocator{client: client, pr: pr, getReservedPorts: getReservedPorts}
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
	alloc.pa, err = portallocator.NewPortAllocatorCustom(*alloc.pr, func(max int, rangeSpec string) (allocator.Interface, error) {
		return allocator.NewContiguousAllocationMap(max, rangeSpec), nil
	})
	if err != nil {
		return err
	}

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

	return nil
}

// GetAvailablePorts requests portNum ports from the port allocator.
// It returns an int array with allocated ports in it.
func (alloc *RuntimePortAllocator) GetAvailablePorts(portNum int) (ports []int, err error) {
	if alloc.pa == nil {
		return nil, errors.New("Runtime port allocator not setup")
	}

	for i := 0; i < portNum; i++ {
		if availPort, err := alloc.pa.AllocateNext(); err != nil {
			alloc.log.Error(err, "can't allocate next, all ports are in use")
			break
		} else {
			ports = append(ports, availPort)
		}
	}

	// Something unexpected happened, rollback to release allocated ports
	if len(ports) < portNum {
		for _, reservedPort := range ports {
			_ = alloc.pa.Release(reservedPort)
		}
		return nil, errors.Errorf("can't get enough available ports, only %d ports are available", len(ports))
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
