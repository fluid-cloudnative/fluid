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
	"testing"

	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var dummy = func(client client.Client) (ports []int, err error) {
	return []int{20001, 20002, 20003}, nil
}

var errDummy = func(client client.Client) (ports []int, err error) {
	return nil, errors.New("err")
}

func TestRuntimePortAllocatorWithError(t *testing.T) {
	pr := net.ParsePortRangeOrDie("20000-21000")
	err := SetupRuntimePortAllocator(nil, pr, "bitmap", errDummy)
	if err != nil {
		t.Fatalf("failed to setup runtime port allocator due to %v", err)
	}

	_, err = GetRuntimePortAllocator()
	if err == nil {
		t.Errorf("Expecetd error when GetRuntimePortAllocator")
	}
}

func TestRuntimePortAllocator(t *testing.T) {
	pr := net.ParsePortRangeOrDie("20000-21000")
	err := SetupRuntimePortAllocator(nil, pr, "bitmap", dummy)
	if err != nil {
		t.Errorf("get non-nil err when GetRuntimePortAllocator")
		return
	}

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

func TestRuntimePortAllocatorRelease(t *testing.T) {
	pr := net.ParsePortRangeOrDie("20000-20010")
	err := SetupRuntimePortAllocator(nil, pr, "bitmap", dummy)
	if err != nil {
		t.Errorf("get non-nil err when GetRuntimePortAllocator")
		return
	}

	preservedPorts, _ := dummy(nil)

	allocator, err := GetRuntimePortAllocator()
	if err != nil {
		t.Errorf("get non-nil err when GetRuntimePortAllocator")
		return
	}

	allocatedPorts, err := allocator.GetAvailablePorts(pr.Size - len(preservedPorts))

	if err != nil || containsAny(allocatedPorts, preservedPorts) {
		t.Errorf("get non-nil err when GetAvailablePortAllocator")
		return
	}

}

func containsAny(ports []int, dst []int) bool {
	m := map[int]bool{}
	for _, v := range ports {
		m[v] = true
	}
	for _, v := range dst {
		_, ok := m[v]
		if ok {
			return true
		}
	}

	return false
}

func TestUnknownPortAllocator(t *testing.T) {
	pr := net.ParsePortRangeOrDie("1000-1100")
	SetupRuntimePortAllocatorWithType(nil, pr, "unknown", dummy)

	_, err := GetRuntimePortAllocator()
	if err == nil {
		t.Errorf("get non-nil err when GetRuntimePortAllocator")
		return
	}
}

func TestRandomRuntimePortAllocator(t *testing.T) {
	pr := net.ParsePortRangeOrDie("1000-1100")
	SetupRuntimePortAllocatorWithType(nil, pr, Random, dummy)

	allocator, err := GetRuntimePortAllocator()
	if err != nil {
		t.Errorf("get non-nil err when GetRuntimePortAllocator")
		return
	}

	_, err = allocator.GetAvailablePorts(pr.Size + 1)
	if err == nil {
		t.Errorf("allocate ports shoule have error")
		return
	}

	allocatedPorts, err := allocator.GetAvailablePorts(pr.Size)
	if err != nil {
		t.Errorf("get non-nil err when GetAvailablePortAllocator")
		return
	}
	if len(allocatedPorts) != pr.Size {
		t.Errorf("allocate ports size less than required")
		return
	}
	if !between(allocatedPorts, pr.Base, pr.Base+pr.Size) || hasDuplicatedElement(allocatedPorts) {
		t.Errorf("allocate ports are not all valid")
		return
	}

	toRelease := []int{20003, 20004}
	allocator.ReleaseReservedPorts(toRelease)

}

func hasDuplicatedElement(ports []int) bool {
	m := map[int]bool{}
	for _, v := range ports {
		m[v] = true
	}
	return len(m) != len(ports)
}

func between(a []int, min int, max int) bool {
	for _, value := range a {
		if value < min && value > max {
			return false
		}
	}
	return true
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
