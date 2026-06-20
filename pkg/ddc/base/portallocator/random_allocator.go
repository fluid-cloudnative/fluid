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

// Allocate is a no-op for RandomAllocator.
// Individual port allocation checks are not currently required by this allocator.
func (r *RandomAllocator) Allocate(port int) error {
	// not judge whether port can be allocated or not
	return nil

}

// Release releases the resource associated with the given index.
// For the RandomAllocator, no actual resource is allocated per index,
// so this method does nothing and always returns a nil error.
//
// Parameters:
//   - i: the index of the resource to release (unused).
//
// Returns:
//   - error: always nil, indicating success with no action.
func (r *RandomAllocator) Release(i int) error {
	// no need to release
	return nil
}

// AllocateBatch attempts to allocate a batch of unique random ports from the allocator's port range.
// The number of ports to allocate is specified by portNum.
// It returns a slice of allocated port numbers and an error if allocation fails.
// An error is returned when portNum exceeds the total size of the port range.
// This method is concurrency-safe as it acquires a lock before modifying the port allocation state.
// Note: The allocated ports are not persisted; they are only stored in the local map to avoid duplicates within the same batch.
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
