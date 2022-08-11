package portallocator

import (
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
}

func newRandomAllocator(pr *net.PortRange) *RandomAllocator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &RandomAllocator{
		portRange: pr,
		rand:      r,
	}
}

func (r *RandomAllocator) Allocate(port int) error {
	// not judge whether port can be allocated or not
	return nil

}

func (r *RandomAllocator) AllocateNext() (int, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.portRange.Base + r.rand.Intn(r.portRange.Size), nil
}

func (r *RandomAllocator) Release(i int) error {
	// no need to release
	return nil
}

func (r *RandomAllocator) ForEach(f func(int)) {
	// not used
	panic("implement me")
}

func (r *RandomAllocator) Has(i int) bool {
	// not used
	panic("implement me")
}
