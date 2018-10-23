package pair

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newPair() *pair { return New().(*pair) }

type mockEP struct {
	inq, outq chan interface{}
	dq        chan struct{}
}

func newMockEP() *mockEP {
	return &mockEP{
		inq:  make(chan interface{}, 1),
		outq: make(chan interface{}, 1),
		dq:   make(chan struct{}, 1),
	}
}

func (m mockEP) Done() <-chan struct{}      { return m.dq }
func (m mockEP) Inbox() <-chan interface{}  { return m.inq }
func (m mockEP) Outbox() chan<- interface{} { return m.outq }

func TestNewPair(t *testing.T) {
	assert.NotNil(t, newPair().ready, "ready chan not initialized")
}

func TestEndpoints(t *testing.T) {
	pair := newPair()
	ep0 := newMockEP()
	ep1 := newMockEP()

	t.Run("Add", func(t *testing.T) {
		t.Run("FirstEndpoint", func(t *testing.T) {
			pair.AddEndpoint(ep0)
			assert.NotNil(t, pair.left, "left endpoint not set")
			assert.Equal(t, pair.left, ep0, "wrong object in left endpoint slot")
			assert.Nil(t, pair.right, "right endpoint slot unexpectedly set")
		})

		t.Run("SecondEndpoint", func(t *testing.T) {
			pair.AddEndpoint(ep1)
			assert.NotNil(t, pair.right, "left endpoint not set")
			assert.Equal(t, pair.left, ep0, "wrong object in left endpoint slot")
			assert.Equal(t, pair.right, ep1, "wrong object in right endpoint slot")

			select {
			case <-pair.ready:
			default:
				t.Error("endpoints set but protocol not ready")
			}
		})

		t.Run("ExcessiveEndpoint", func(t *testing.T) {
			assert.Panics(t, func() { pair.AddEndpoint(newMockEP()) })
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("Left", func(t *testing.T) {
			pair.RemoveEndpoint(ep0)
			assert.Nil(t, pair.left, "left endpoint not cleared")
			assert.NotNil(t, pair.right, "right endpoint erroneously cleared")
		})

		// Make sure that we can continue if a new endpoint is added
		t.Run("ReplaceLeft", func(t *testing.T) {
			pair.AddEndpoint(ep0)
			select {
			case <-pair.ready:
			default:
				t.Error("endpoints set but protocol not ready")
			}
		})

		t.Run("Right", func(t *testing.T) {
			pair.RemoveEndpoint(ep1)
			assert.Nil(t, pair.right, "right endpoint not cleared")
			assert.NotNil(t, pair.left, "left endpoint erroneously cleared")
		})

		// Make sure that we can continue if a new endpoint is added
		t.Run("ReplaceRight", func(t *testing.T) {
			pair.AddEndpoint(ep1)
			select {
			case <-pair.ready:
			default:
				t.Error("endpoints set but protocol not ready")
			}
		})
	})
}

func TestRelayMessages(t *testing.T) {
	// pair := newPair()
	// ep0 := newMockEP()
	// ep1 := newMockEP()

	// // messages to be sent
	// ep0.outq <- 0
	// ep1.outq <- 1

	// // add the endpoints & start the protocol
	// pair.AddEndpoint(ep0)
	// pair.AddEndpoint(ep1)
	// go pair.Init(sigctx.New())

	// t.Run("LeftToRight", func(t *testing.T) {
	// 	assert.Equal(t, <-ep0.Inbox(), 1)
	// })

	// t.Run("RightToLeft", func(t *testing.T) {
	// 	assert.Equal(t, <-ep1.Inbox(), 0)
	// })
}
