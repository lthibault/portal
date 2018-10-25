package portal

// OptAsync creates a buffered chan of the specified size
func OptAsync(bufSize int) func(*portal) {
	return func(p *portal) {
		p.bufSize = bufSize
	}
}
