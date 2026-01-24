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
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var dummy = func(client client.Client) (ports []int, err error) {
	return []int{20001, 20002, 20003}, nil
}

var errDummy = func(client client.Client) (ports []int, err error) {
	return nil, errors.New("err")
}

var _ = Describe("RuntimePortAllocator", func() {
	BeforeEach(func() {
		rpa = nil
	})

	Context("when setup with error", func() {
		It("should return error when getting allocator", func() {
			pr := net.ParsePortRangeOrDie("20000-21000")
			err := SetupRuntimePortAllocator(nil, pr, "bitmap", errDummy)
			Expect(err).NotTo(HaveOccurred())

			_, err = GetRuntimePortAllocator()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when releasing ports", func() {
		It("should not allocate preserved ports", func() {
			pr := net.ParsePortRangeOrDie("20000-20010")
			err := SetupRuntimePortAllocator(nil, pr, "bitmap", dummy)
			Expect(err).NotTo(HaveOccurred())

			preservedPorts, _ := dummy(nil)

			allocator, err := GetRuntimePortAllocator()
			Expect(err).NotTo(HaveOccurred())

			allocatedPorts, err := allocator.GetAvailablePorts(pr.Size - len(preservedPorts))
			Expect(err).NotTo(HaveOccurred())
			Expect(containsAny(allocatedPorts, preservedPorts)).To(BeFalse())
		})

		It("should make released ports available for re-allocation", func() {
			pr := net.ParsePortRangeOrDie("20000-20005")
			err := SetupRuntimePortAllocator(nil, pr, "bitmap", dummy)
			Expect(err).NotTo(HaveOccurred())

			preservedPorts, _ := dummy(nil)
			allocator, err := GetRuntimePortAllocator()
			Expect(err).NotTo(HaveOccurred())

			// Allocate all non-preserved ports (range has 6 ports, 3 are preserved, so 3 available)
			firstAllocation, err := allocator.GetAvailablePorts(pr.Size - len(preservedPorts))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(firstAllocation)).To(Equal(3))

			// Release 2 ports
			portsToRelease := firstAllocation[:2]
			allocator.ReleaseReservedPorts(portsToRelease)

			// Now allocate 2 more ports - these MUST include the released ports
			// since we only have 1 unreleased port left
			secondAllocation, err := allocator.GetAvailablePorts(2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(secondAllocation)).To(Equal(2))

			// At least one of the released ports must be in the second allocation
			hasReleasedPort := containsAny(secondAllocation, portsToRelease)
			Expect(hasReleasedPort).To(BeTrue(), "At least one released port should be re-allocated")
		})
	})
})

var _ = Describe("UnknownPortAllocator", func() {
	It("should return error for unknown allocator type", func() {
		pr := net.ParsePortRangeOrDie("1000-1100")
		SetupRuntimePortAllocatorWithType(nil, pr, "unknown", dummy)

		_, err := GetRuntimePortAllocator()
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("RandomRuntimePortAllocator", func() {
	var pr *net.PortRange
	var allocator *RuntimePortAllocator

	BeforeEach(func() {
		pr = net.ParsePortRangeOrDie("1000-1100")
		SetupRuntimePortAllocatorWithType(nil, pr, Random, dummy)

		var err error
		allocator, err = GetRuntimePortAllocator()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return error when allocating more ports than available", func() {
		_, err := allocator.GetAvailablePorts(pr.Size + 1)
		Expect(err).To(HaveOccurred())
	})

	It("should allocate all available ports successfully", func() {
		allocatedPorts, err := allocator.GetAvailablePorts(pr.Size)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(allocatedPorts)).To(Equal(pr.Size))
		Expect(between(allocatedPorts, pr.Base, pr.Base+pr.Size)).To(BeTrue())
		Expect(hasDuplicatedElement(allocatedPorts)).To(BeFalse())
	})

	It("should release reserved ports", func() {
		toRelease := []int{20003, 20004}
		allocator.ReleaseReservedPorts(toRelease)
	})
})

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

func hasDuplicatedElement(ports []int) bool {
	m := map[int]bool{}
	for _, v := range ports {
		m[v] = true
	}
	return len(m) != len(ports)
}

func between(a []int, min int, max int) bool {
	for _, value := range a {
		if value < min || value >= max {
			return false
		}
	}
	return true
}
