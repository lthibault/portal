package pair

import (
	"testing"
	"time"

	"github.com/SentimensRG/ctx"

	"github.com/SentimensRG/ctx/sigctx"
	"github.com/lthibault/portal"
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

func (mockEP) Close()                       {}
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

func TestIntegration(t *testing.T) {
	d, cancel := ctx.WithCancel(sigctx.New())
	defer cancel()

	p := portal.New(New(), portal.OptCtx(d))

	ch0 := p.Open()
	ch1 := p.Open()
	defer ch0.Close()
	defer ch1.Close()

	go func() { ch0.Send() <- 0 }()
	go func() { ch1.Send() <- 1 }()

	select {
	case i := <-ch0.Recv():
		assert.Equal(t, i.(int), 1)
	case <-time.After(time.Millisecond * 10):
		t.Error("ch0 failed to receive")
	}

	select {
	case i := <-ch1.Recv():
		assert.Equal(t, i.(int), 0)
	case <-time.After(time.Millisecond * 10):
		t.Error("ch1 failed to receive")
	}

}
