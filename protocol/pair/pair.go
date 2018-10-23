package pair

import (
	"errors"
	"sync"

	"github.com/SentimensRG/ctx"
	"github.com/lthibault/portal"
)

type pair struct {
	sync.Mutex
	ready       chan struct{}
	left, right portal.Endpoint
}

// New pair protocol
func New() portal.Protocol {
	return &pair{ready: make(chan struct{}, 1)}
}

func (p *pair) Init(d ctx.Doner) {
	for {
		select {
		case <-d.Done():
			return
		case <-p.ready:
			p.run(d)
		}
	}
}

func (p *pair) run(d ctx.Doner) {
	go p.relay(p.left, p.right)
	go p.relay(p.right, p.left)
}

func (p *pair) relay(src, dst portal.Endpoint) {
	for {
		select {
		case v := <-src.Inbox():
			select {
			case dst.Outbox() <- v:
			case <-dst.Done():
				return
			}
		case <-src.Done():
			return
		}
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
	} else {
		panic(errors.New("no such endpoint"))
	}
}
