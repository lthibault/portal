package portal

import (
	"errors"
	"runtime"
)

// Portal provides a thread-safe communication channel with messaging patterns.
type Portal interface {
	Open() Chan
	Close()
}

// Chan is a two-way channel
type Chan interface {
	Send() chan<- interface{}
	Recv() <-chan interface{}
	Close()
}

// ErrChanClosed is thrown in a panic if a Chan is closed more than once
var ErrChanClosed = errors.New("close of closed portal.Chan")

// Protocol defines a messaging topology
type Protocol interface {
	Close()
	AddEndpoint(Endpoint)
	RemoveEndpoint(Endpoint)
}

// Endpoint represents the different IO channels that a Protocol must manage
type Endpoint interface {
	Close()
	Done() <-chan struct{}
	Inbox() <-chan interface{}
	Outbox() chan<- interface{}
}

type option func(*portal)

type endpointRemover interface {
	RemoveEndpoint(Endpoint)
}

func checkDoubleClose() {
	if r := recover(); r != nil {
		if _, ok := r.(runtime.Error); ok {
			r = ErrChanClosed
		}
		panic(r)
	}
}

// pipe implements chan & endpoint
type pipe struct {
	cq           chan struct{}
	sendq, recvq chan interface{}
	epRm         endpointRemover
}

// Close the pipe
func (p pipe) Close() {
	defer checkDoubleClose()

	close(p.cq)
	p.epRm.RemoveEndpoint(p)
	close(p.sendq)
	close(p.recvq)
}

func (p pipe) Done() <-chan struct{} { return p.cq }

// Send pipe
func (p pipe) Send() chan<- interface{}  { return p.sendq }
func (p pipe) Inbox() <-chan interface{} { return p.sendq }

// Recv pipe
func (p pipe) Recv() <-chan interface{}   { return p.recvq }
func (p pipe) Outbox() chan<- interface{} { return p.recvq }

type portal struct {
	proto   Protocol
	bufSize int
}

func (p *portal) Open() Chan {
	pp := &pipe{
		cq:    make(chan struct{}),
		sendq: make(chan interface{}, p.bufSize),
		recvq: make(chan interface{}, p.bufSize),
		epRm:  p.proto,
	}

	p.proto.AddEndpoint(pp)
	return pp
}

func (p *portal) Close() {
	defer checkDoubleClose()
	p.proto.Close()
}

// New Portal
func New(p Protocol, opts ...option) Portal {
	ptl := &portal{proto: p}

	for _, opt := range opts {
		opt(ptl)
	}

	return ptl
}
