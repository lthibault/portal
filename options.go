package portal

import "github.com/SentimensRG/ctx"

// OptAsync creates a buffered chan of the specified size
func OptAsync(bufSize int) func(*portal) {
	return func(p *portal) {
		p.bufSize = bufSize
	}
}

// OptCtx sets the root context for the Portal.  When the context expires,
// the Portal will be closed, along with all Chans.
func OptCtx(d ctx.Doner) func(*portal) {
	return func(p *portal) {
		p.d, p.closer = ctx.WithCancel(d)
		return
	}
}
