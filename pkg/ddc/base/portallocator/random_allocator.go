package portallocator

import (
	"errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/net"
	"math/rand"
	"sync"
	"time"
)

type RandomAllocator struct {
	portRange *net.PortRange
	// lock make rand thread safe
	lock sync.Mutex
	rand *rand.Rand
	log  logr.Logger
}

func newRandomAllocator(pr *net.PortRange, log logr.Logger) *RandomAllocator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &RandomAllocator{
		portRange: pr,
		rand:      r,
		log:       log,
	}
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
