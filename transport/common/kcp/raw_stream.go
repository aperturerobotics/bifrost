package kcp

import (
	"context"
	"sync"

	"github.com/aperturerobotics/bifrost/stream"
	"github.com/djherbis/buffer"
	"github.com/golang/protobuf/proto"
)

// rawStream implements a unencrypted stream.
type rawStream struct {
	*PacketBuffer

	// ctx is the stream context
	ctx context.Context

	packetMtx  sync.Mutex
	packetBuf  buffer.Buffer
	packetWake chan struct{}

	closeFn     func(*rawStream)
	establishCb func(err error)

	// localStreamID is the local stream id
	localStreamID uint32

	// closed indicates this stream is already disposed
	// guarded by the rawStreamMtx in the link
	closed bool
	// closeFnOnce calls closeFn once
	closeFnOnce sync.Once
	// mtu is the maximum transmission unit
	mtu      uint32
	writeMtx sync.Mutex
}

func revSlice(vb []byte) {
	for i := len(vb)/2 - 1; i >= 0; i-- {
		opp := len(vb) - 1 - i
		vb[i], vb[opp] = vb[opp], vb[i]
	}
}

func encodeRawStreamIDVarint(streamID uint32) []byte {
	vb := proto.EncodeVarint(uint64(streamID))
	// reverse array
	revSlice(vb)
	return vb
}

// decodeRawStreamIDVarint temporarily reverses the bytes and decodes the varint.
func decodeRawStreamIDVarint(vb []byte) (x uint64, n int) {
	revSlice(vb)
	x, n = proto.DecodeVarint(vb)
	revSlice(vb)
	return
}

func newRawStream(
	ctx context.Context,
	localStreamID uint32,
	mtu uint32,
	establishCb func(err error),
	writeFn func(data []byte) error,
	closeFn func(*rawStream),
) *rawStream {
	return &rawStream{
		PacketBuffer: NewPacketBuffer(ctx, writeFn),

		ctx:           ctx,
		establishCb:   establishCb,
		closeFn:       closeFn,
		mtu:           mtu,
		localStreamID: localStreamID,
	}
}

// markClosed marks the stream as closed.
func (s *rawStream) markClosed() {
	if s.closed {
		return
	}

	s.closed = true
}

// Write data to the stream.
func (s *rawStream) Write(b []byte) (n int, err error) {
	var nw int
	blen := len(b)
	mtu := int(s.mtu) - len(s.headerVarint)

	s.writeMtx.Lock()
	defer s.writeMtx.Unlock()

	// If fragmentation is necessary...
	// slow path
	// TODO: this may incorrectly be re-assembled on the other end
	// TODO: maybe use native IP fragmenting here instead.
	if blen > mtu {
		var xpkt []byte
		for i := 0; i < blen; i += mtu {
			j := i + mtu
			if j > blen {
				j = blen
			}

			jWithSuffix := j + len(s.headerVarint)
			var pkt []byte
			if j == blen && jWithSuffix < cap(b) {
				// if we know we will re-use the end of the b buffer
				// and it's safe because j is past the end of the buff
				// but still within capacity
				pkt = b[i:jWithSuffix]
			} else {
				if xpkt != nil {
					pkt = xpkt
				} else {
					xpkt = make([]byte, s.mtu)
					pkt = xpkt
				}
				pkt = pkt[:jWithSuffix-i]
				copy(pkt, b[i:j])
			}
			copy(pkt[j-i:], s.headerVarint)

			if _, err := s.PacketBuffer.Write(pkt); err != nil {
				return nw, err
			}
			nw += jWithSuffix - i
		}

		return nw, nil
	}

	return s.PacketBuffer.Write(b)
}

// Close closes the stream.
func (s *rawStream) Close() error {
	s.closeFnOnce.Do(func() {
		go s.closeFn(s)
	})
	return nil
}

// _ is a type assertion
var _ stream.Stream = ((*rawStream)(nil))
