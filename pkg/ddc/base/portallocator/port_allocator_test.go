package portallocator

import (
	"errors"
	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

var dummy = func(client client.Client) (ports []int, err error) {
	return []int{20001, 20002, 20003}, nil
}

var errDummy = func(client client.Client) (ports []int, err error) {
	return nil, errors.New("err")
}

func TestRuntimePortAllocatorWithError(t *testing.T) {
	pr := net.ParsePortRangeOrDie("20000-21000")
	SetupRuntimePortAllocator(nil, pr, errDummy)

	_, err := GetRuntimePortAllocator()
	if err == nil {
		t.Errorf("Expecetd error when GetRuntimePortAllocator")
	}
}

func TestRuntimePortAllocator(t *testing.T) {
	pr := net.ParsePortRangeOrDie("20000-21000")
	SetupRuntimePortAllocator(nil, pr, dummy)

	allocator, err := GetRuntimePortAllocator()
	if err != nil {
		t.Errorf("get non-nil err when GetRuntimePortAllocator")
		return
	}

	expected := []int{20004, 20005, 20006}
	allocatedPorts, err := allocator.GetAvailablePorts(3)
	if err != nil || sameArray(expected, allocatedPorts) {
		t.Errorf("get non-nil err when GetAvailablePortAllocator")
		return
	}

	toRelease := []int{20003, 20004}
	allocator.ReleaseReservedPorts(toRelease)

	expected = []int{20003, 20004, 20007, 20008}
	allocatedPorts, err = allocator.GetAvailablePorts(4)
	if err != nil || sameArray(expected, allocatedPorts) {
		t.Errorf("get non-nil err when GetAvailablePortAllocator")
		return
	}
}

func sameArray(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	lenArr := len(a)
	for i := 0; i < lenArr; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
