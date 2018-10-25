package test

// EP is a mock endpoint
type EP struct {
	InQ, OutQ chan interface{}
	DoneQ     chan struct{}
}

// NewEP returns a mock portal.Endpoint
func NewEP() *EP {
	return &EP{
		DoneQ: make(chan struct{}),
		InQ:   make(chan interface{}, 1),
		OutQ:  make(chan interface{}, 1),
	}
}

// Close the Endpoint
func (m EP) Close() {
	close(m.DoneQ)
	close(m.InQ)
	close(m.OutQ)
}

// Done indicates the Endpoint is closed
func (m EP) Done() <-chan struct{} { return m.DoneQ }

// Inbox is a queue of messages received
func (m EP) Inbox() <-chan interface{} { return m.InQ }

// Outbox is a queue of messages sent
func (m EP) Outbox() chan<- interface{} { return m.OutQ }
