package proxy

import (
	"github.com/aperturerobotics/bifrost/stream"
	"io"
)

func proxyStreamTo(s1, s2 stream.Stream) {
	buf := make([]byte, 10000)
	io.CopyBuffer(s2, s1, buf)
	// io.Copy(s2, s1)
	s1.Close()
	s2.Close()
}

// ProxyStreams constructs read/write pumps to proxy two streams.
// if either stream is closed, the other will be closed.
// The two routines will exit when the streams are closed.
func ProxyStreams(s1, s2 stream.Stream) {
	go proxyStreamTo(s1, s2)
	go proxyStreamTo(s2, s1)
}
