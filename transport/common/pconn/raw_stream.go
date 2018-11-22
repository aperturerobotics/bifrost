package pconn

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/stream"
	"github.com/golang/protobuf/proto"
)

// rawStream implements a unencrypted stream.
type rawStream struct {
	// ctx is the stream context
	ctx context.Context
	// localStreamID is the local stream id
	localStreamID uint32
	// remoteStreamID is the remote stream id
	remoteStreamID uint32
	// headerVarint is the header varint
	// the bytes are reversed.
	// empty until remote stream ID is known
	headerVarint []byte

	// packetCh is a packet buffer implemented with a channel
	packetCh    chan []byte
	writeFn     func(data []byte) error
	writeMtx    sync.Mutex
	closeFn     func(*rawStream)
	establishCb func(err error)

	readDeadline, writeDeadline time.Time

	// closed indicates this stream is already disposed
	// guarded by the rawStreamMtx in the link
	closed bool
	// mtu is the maximum transmission unit
	mtu uint32
	// closeFnOnce calls closeFn once
	closeFnOnce sync.Once
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
		ctx: ctx,
		// keep 60000 packets
		// 1 packet = 1500 byte maximum
		// 60000*1500 ~= 100mb packet ring
		// can be replaced with a better ringbuffer impl later
		packetCh:      make(chan []byte, mtu*60000),
		establishCb:   establishCb,
		writeFn:       writeFn,
		closeFn:       closeFn,
		mtu:           mtu,
		localStreamID: localStreamID,
	}
}

// SetRemoteStreamID sets the remote stream ID varint trailer
func (s *rawStream) SetRemoteStreamID(id uint32) {
	s.remoteStreamID = id
	s.headerVarint = encodeRawStreamIDVarint(id)
}

// PushPacket pushes a packet to the stream, dropping the oldest packet if necessary.
func (s *rawStream) PushPacket(packet []byte) {
	if s.closed {
		return
	}

PushLoop:
	for {
		select {
		case s.packetCh <- packet:
			break PushLoop
		default:
		}

		// drop a packet
		select {
		case pkt := <-s.packetCh:
			xmitBuf.Put(pkt)
		default:
		}
	}
}

// Read data from the stream.
func (s *rawStream) Read(b []byte) (n int, err error) {
	readDeadline := s.readDeadline
	var c <-chan time.Time
	if !readDeadline.IsZero() {
		now := time.Now()
		if readDeadline.Before(now) {
			return 0, context.DeadlineExceeded
		}

		fromNow := readDeadline.Sub(now)
		tck := time.NewTicker(fromNow)
		defer tck.Stop()
		c = tck.C
	}

	select {
	case <-s.ctx.Done():
		return 0, io.EOF
	case pkt, ok := <-s.packetCh:
		if !ok {
			return 0, io.EOF
		}

		copy(b, pkt)
		xmitBuf.Put(pkt)
		if len(b) < len(pkt) {
			return len(b), io.ErrShortBuffer
		}

		return len(pkt), nil
	case <-c:
		return 0, context.DeadlineExceeded
	}
}

// Write data to the stream.
func (s *rawStream) Write(b []byte) (n int, err error) {
	// TODO: Fragment here.
	var nw int
	mtu := int(s.mtu) - len(s.headerVarint)
	blen := len(b)

	s.writeMtx.Lock()
	defer s.writeMtx.Unlock()

	// If fragmentation is necessary...
	// slow path
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
					pkt = xmitBuf.Get().([]byte)
					xpkt = pkt
					defer func() {
						xmitBuf.Put(xpkt)
					}()
				}
				pkt = pkt[:jWithSuffix-i]
				copy(pkt, b[i:j])
			}
			copy(pkt[j-i:], s.headerVarint)

			if err := s.writeFn(pkt); err != nil {
				return nw, err
			}
			nw += jWithSuffix - i
		}

		return nw, nil
	}

	blenWithSuffix := blen + len(s.headerVarint)
	if cap(b) < blenWithSuffix {
		b = append(b, s.headerVarint...)
	} else {
		b = b[:blenWithSuffix]
		copy(b[blen:], s.headerVarint)
	}

	if err := s.writeFn(b); err != nil {
		return 0, err
	}

	return blen, nil
}

// SetReadDeadline sets the read deadline as defined by
// A zero time value disables the deadline.
func (s *rawStream) SetReadDeadline(t time.Time) error {
	s.readDeadline = t
	return nil
}

// SetWriteDeadline sets the write deadline as defined by
// A zero time value disables the deadline.
func (s *rawStream) SetWriteDeadline(t time.Time) error {
	s.writeDeadline = t
	return nil
}

// SetDeadline sets both read and write deadlines as defined by
// A zero time value disables the deadlines.
func (s *rawStream) SetDeadline(t time.Time) error {
	s.readDeadline = t
	s.writeDeadline = t
	return nil
}

// markClosed marks the stream as closed.
func (s *rawStream) markClosed() {
	if s.closed {
		return
	}

	s.closed = true
	close(s.packetCh)
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
