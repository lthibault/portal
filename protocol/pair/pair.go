package pair

import (
	"errors"
	"sync"

	"github.com/lthibault/portal"
)

type pair struct {
	sync.RWMutex
	ready       chan struct{}
	left, right portal.Endpoint
}

func newPair() *pair { return &pair{ready: make(chan struct{}, 1)} }

// New pair protocol
func New() portal.Protocol {
	p := newPair()
	go p.init()
	return p
}

func (p *pair) init() {
	for range p.ready {
		go p.relay(p.left, p.right)
		go p.relay(p.right, p.left)
	}
}

func (p *pair) relay(src, dst portal.Endpoint) {
	defer recover()
	for v := range src.Inbox() {
		func() {
			p.RLock()
			defer p.RUnlock()
			dst.Outbox() <- v
		}()
	}
}

func (p *pair) Close() {
	close(p.ready)
	if p.left != nil {
		p.left.Close()
	}
	if p.right != nil {
		p.right.Close()
	}
}

func (p *pair) AddEndpoint(ep portal.Endpoint) {
	p.Lock()
	defer p.Unlock()

	if p.left == nil {
		p.left = ep
	} else if p.right == nil {
		p.right = ep
	} else {
		panic(errors.New("pair supports exactly two endpoints"))
	}

	if p.left != nil && p.right != nil {
		p.ready <- struct{}{}
	}
}

func (p *pair) RemoveEndpoint(ep portal.Endpoint) {
	p.Lock()
	defer p.Unlock()

	if ep == p.left {
		p.left = nil
	} else if ep == p.right {
		p.right = nil
	}
}
