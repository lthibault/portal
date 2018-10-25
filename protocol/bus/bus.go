package bus

import (
	"sync"

	"github.com/SentimensRG/ctx"
	"github.com/lthibault/portal"
)

type bus struct {
	sync.RWMutex
	eps map[portal.Endpoint]struct{}
}

// New bus protocol
func New() portal.Protocol {
	return &bus{eps: make(map[portal.Endpoint]struct{})}
}

func (b *bus) Init(_ ctx.Doner) {}

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
			// defer recover() // chan send may panic
			defer wg.Done()

			select {
			case <-ep.Done():
			case ep.Outbox() <- v:
			}

		}(ep)
	}

	wg.Wait()
}

func (b *bus) AddEndpoint(ep portal.Endpoint) {
	b.Lock()
	b.eps[ep] = struct{}{}
	b.Unlock()
	go func() {
		for v := range ep.Inbox() {
			b.broadcast(v, ep)
		}
	}()
}

func (b *bus) RemoveEndpoint(ep portal.Endpoint) {
	b.Lock()
	delete(b.eps, ep)
	b.Unlock()
}
