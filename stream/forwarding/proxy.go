package stream_forwarding

import (
	"github.com/aperturerobotics/bifrost/stream"
	"io"
)

func proxyStreamTo(s1, s2 stream.Stream) {
	io.Copy(s2, s1)
	s2.Close()
	s1.Close()
}

// ProxyStreams constructs read/write pumps to proxy two streams.
// if either stream is closed, the other will be closed.
// The two routines will exit when the streams are closed.
func ProxyStreams(s1, s2 stream.Stream) {
	go proxyStreamTo(s1, s2)
	go proxyStreamTo(s2, s1)
}
