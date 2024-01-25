package kcp

import (
	"sync"
)

// global packet buffer
// shared among sending/receiving
var xmitBuf sync.Pool

func init() {
	xmitBuf.New = func() interface{} {
		return make([]byte, 1500)
	}
}
