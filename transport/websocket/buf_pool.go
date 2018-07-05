package websocket

import (
	"sync"
)

var (
	// global packet buffer
	// shared among sending/receiving
	xmitBuf sync.Pool
)

func init() {
	xmitBuf.New = func() interface{} {
		return make([]byte, 1500)
	}
}
