package scrub

// Scrub clears a buffer with zeros.
// Prevents reading sensitive data before memory is overwritten.
func Scrub(buf []byte) {
	// compiler optimizes this to memset
	for i := 0; i < len(buf); i++ {
		buf[i] = 0
	}
}
