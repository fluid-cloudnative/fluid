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

type RuntimePortAllocator struct {
	pa     *portallocator.PortAllocator
	client client.Client
	pr     *net.PortRange

	getReservedPorts func(client client.Client) (ports []int, err error)

	log logr.Logger
}

var rpa *RuntimePortAllocator

func SetupRuntimePortAllocator(client client.Client, pr *net.PortRange, getReservedPorts func(client client.Client) (ports []int, err error)) {
	rpa = &RuntimePortAllocator{client: client, pr: pr, getReservedPorts: getReservedPorts}
	rpa.log = ctrl.Log.WithName("RuntimePortAllocator")
}

func GetRuntimePortAllocator() (*RuntimePortAllocator, error) {
	if rpa.pa == nil {
		if err := rpa.createAndRestorePortAllocator(); err != nil {
			return nil, err
		}
	}
	return rpa, nil
}

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

func (alloc *RuntimePortAllocator) GetAvailablePorts(portNum int) (ports []int, err error) {
	if alloc.pa == nil {
		return nil, errors.Errorf("Runtime port allocator not setup")
	}

	for i := 0; i < portNum; i++ {
		if availPort, err := alloc.pa.AllocateNext(); err != nil {
			alloc.log.Error(err, "can't allocate next")
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
		return nil, errors.Errorf("can't get available ports, only %d ports are available", len(ports))
	}

	alloc.log.Info("GetAvailablePorts", "expecetedPortNum", portNum, "gotPorts", ports)
	return ports, nil
}

func (alloc *RuntimePortAllocator) ReleaseReservedPorts(ports []int) {
	for _, port := range ports {
		err := alloc.pa.Release(port)
		if err != nil {
			alloc.log.Info("can't release port, ignore it", "port", port)
		}
	}
}
