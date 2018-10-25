package test

// EP is a mock endpoint
type EP struct {
	InQ, OutQ chan interface{}
	DoneQ     chan struct{}
}

func NewEP() *EP {
	return &EP{
		InQ:   make(chan interface{}, 1),
		OutQ:  make(chan interface{}, 1),
		DoneQ: make(chan struct{}, 1),
	}
}

// Done implements ctx.Doner
func (m EP) Done() <-chan struct{} { return m.DoneQ }

// Inbox is a queue of messages received
func (m EP) Inbox() <-chan interface{} { return m.InQ }

// Outbox is a queue of messages sent
func (m EP) Outbox() chan<- interface{} { return m.OutQ }
