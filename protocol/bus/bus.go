package bus

import (
	"errors"
	"sync"

	"github.com/lthibault/portal"
)

type bus struct {
	sync.RWMutex
	cq  chan struct{}
	eps map[portal.Endpoint]struct{}
}

// New bus protocol
func New() portal.Protocol {
	return &bus{
		eps: make(map[portal.Endpoint]struct{}),
		cq:  make(chan struct{}),
	}
}

func (b *bus) broadcast(v interface{}, origin portal.Endpoint) {
	b.RLock()
	defer b.RUnlock()

	var wg sync.WaitGroup

	wg.Add(len(b.eps) - 1)

	for ep := range b.eps {
		if ep == origin {
			continue
		}

		go func(ep portal.Endpoint) {
			defer recover() // chan send may panic
			defer wg.Done()

			ep.Outbox() <- v // RACE with concurrent call to Chan.Close()
		}(ep)
	}

	wg.Wait()
}

func (b *bus) Close() {
	close(b.cq)
	for ep := range b.eps {
		select {
		case <-ep.Done():
		default:
			ep.Close()
		}
	}
}

func (b *bus) AddEndpoint(ep portal.Endpoint) {
	select {
	case <-b.cq:
		panic(errors.New("add endpoint to closed protocol"))
	default:
		b.Lock()
		b.eps[ep] = struct{}{}
		b.Unlock()
		go func() {
			for v := range ep.Inbox() {
				b.broadcast(v, ep)
			}
		}()
	}
}

func (b *bus) RemoveEndpoint(ep portal.Endpoint) {
	select {
	case <-b.cq:
	default:
		b.Lock()
		delete(b.eps, ep)
		b.Unlock()
	}
}
