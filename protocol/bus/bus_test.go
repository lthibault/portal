package bus

import (
	"sync"
	"testing"
	"time"

	"github.com/lthibault/portal"

	"github.com/lthibault/portal/test"
	"github.com/stretchr/testify/assert"
)

func TestEndpoints(t *testing.T) {
	bus := New().(*bus)
	ep0 := test.NewEP()
	ep1 := test.NewEP()

	t.Run("Add", func(t *testing.T) {
		t.Run("FirstEndpoint", func(t *testing.T) {
			bus.AddEndpoint(ep0)
			assert.Contains(t, bus.eps, ep0)
		})

		t.Run("SecondEndpoint", func(t *testing.T) {
			bus.AddEndpoint(ep1)
			assert.Contains(t, bus.eps, ep1)
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("Left", func(t *testing.T) {
			bus.RemoveEndpoint(ep0)
			assert.NotContains(t, bus.eps, ep0)
		})

		t.Run("Right", func(t *testing.T) {
			bus.RemoveEndpoint(ep1)
			assert.NotContains(t, bus.eps, ep1)
		})
	})
}

func testFanOut(t *testing.T, pch portal.Chan, chans ...portal.Chan) <-chan struct{} {
	ch := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(len(chans))
	go func() {
		wg.Wait()
		close(ch)
	}()

	datum := struct{}{}
	pch.Send() <- datum

	for _, c := range chans {
		go func(c portal.Chan) {
			if v := <-c.Recv(); v != datum {
				t.Errorf("unexpected value %v", v)
			}
			wg.Done()
		}(c)
	}

	return ch
}

type waitDoner interface {
	Done()
}

func maybeTimeout(t *testing.T, wd waitDoner, ch <-chan struct{}) {
	defer wd.Done()
	select {
	case <-ch:
	case <-time.After(time.Millisecond):
		t.Error("timeout")
	}
}

func TestIntegration(t *testing.T) {
	p := portal.New(New())
	defer p.Close()

	ch0 := p.Open()
	defer ch0.Close()

	ch1 := p.Open()
	defer ch1.Close()

	ch2 := p.Open()
	defer ch2.Close()

	ch3 := p.Open()
	defer ch3.Close()

	var wg sync.WaitGroup
	wg.Add(4)

	go maybeTimeout(t, &wg, testFanOut(t, ch0, ch1, ch2, ch3))
	go maybeTimeout(t, &wg, testFanOut(t, ch1, ch2, ch3, ch0))
	go maybeTimeout(t, &wg, testFanOut(t, ch2, ch3, ch0, ch1))
	go maybeTimeout(t, &wg, testFanOut(t, ch3, ch0, ch1, ch2))

	wg.Wait()
}
