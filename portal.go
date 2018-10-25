package portal

import "github.com/SentimensRG/ctx"

// Portal provides a thread-safe communication channel with messaging patterns.
type Portal interface {
	Open() Chan
	Close()
}

type closeCtx struct {
	ctx.Doner
	closer func()
}

func (c closeCtx) Close() { c.closer() }

// Chan is a two-way channel
type Chan interface {
	ctx.Doner
	Send() chan<- interface{}
	Recv() <-chan interface{}
	Close()
}

// pipe implements chan & endpoint
type pipe struct {
	ctx.Doner
	closer       func()
	sendq, recvq chan interface{}
	epRemover
}

// Close the pipe
func (p pipe) Close() {
	p.closer()
	close(p.sendq)
	close(p.recvq)
}

// Send pipe
func (p pipe) Send() chan<- interface{} { return p.sendq }

func (p pipe) Inbox() <-chan interface{} { return p.sendq }

// Recv pipe
func (p pipe) Recv() <-chan interface{} { return p.recvq }

func (p pipe) Outbox() chan<- interface{} { return p.recvq }

type epRemover interface {
	RemoveEndpoint(Endpoint)
}

type portal struct {
	d       ctx.Doner
	closer  func()
	proto   Protocol
	bufSize int
}

func (p *portal) Open() Chan {
	d, cancel := ctx.WithCancel(p.d)
	pp := &pipe{
		sendq:     make(chan interface{}, p.bufSize),
		recvq:     make(chan interface{}, p.bufSize),
		Doner:     d,
		closer:    cancel,
		epRemover: p.proto,
	}

	p.proto.AddEndpoint(pp)
	ctx.Defer(d, func() { p.proto.RemoveEndpoint(pp) })

	return pp
}

func (p *portal) Close() { p.closer() }

// Protocol defines a messaging topology
type Protocol interface {
	Init(ctx.Doner)
	AddEndpoint(Endpoint)
	RemoveEndpoint(Endpoint)
}

// Endpoint represents the different IO channels that a Protocol must manage
type Endpoint interface {
	ctx.Doner
	Inbox() <-chan interface{}
	Outbox() chan<- interface{}
}

type option func(*portal)

// New Portal
func New(p Protocol, opts ...option) Portal {
	cq := make(chan struct{})
	ptl := &portal{proto: p, d: ctx.C(cq), closer: func() { close(cq) }}
	for _, opt := range opts {
		opt(ptl)
	}

	p.Init(ctx.C(cq))

	return ptl
}
