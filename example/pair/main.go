package main

import (
	"log"
	"time"

	"github.com/SentimensRG/ctx"
	"github.com/SentimensRG/ctx/sigctx"
	"github.com/lthibault/portal"
	"github.com/lthibault/portal/protocol/pair"
)

func main() {
	p := portal.New(pair.New(), portal.OptCtx(sigctx.New()))

	ch0 := p.Open()
	go ctx.FTick(ch0, func() {
		select {
		case v := <-ch0.Recv():
			log.Printf("%v", v)
		case ch0.Send() <- "from ch0":
			time.Sleep(time.Millisecond * 500)
		case <-ch0.Done():
		}
	})

	ch1 := p.Open()
	ctx.FTick(ch0, func() {
		select {
		case v := <-ch1.Recv():
			log.Printf("%v", v)
		case ch1.Send() <- "send to ch1":
			time.Sleep(time.Millisecond * 500)
		case <-ch1.Done():
		}
	})
}
