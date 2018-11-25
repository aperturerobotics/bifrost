package pconn

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/stream"
	"github.com/djherbis/buffer"
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

	packetMtx  sync.Mutex
	packetBuf  buffer.Buffer
	packetWake chan struct{}

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

// keep 6000 packets
// 1 packet = 1500 byte maximum
// 6000*1500 ~= 10mb packet ring
var packetRingSize = 6000 * 1500

func newRawStream(
	ctx context.Context,
	localStreamID uint32,
	mtu uint32,
	establishCb func(err error),
	writeFn func(data []byte) error,
	closeFn func(*rawStream),
) *rawStream {
	return &rawStream{
		ctx:           ctx,
		packetWake:    make(chan struct{}, 1),
		packetBuf:     buffer.NewRing(buffer.New(int64(packetRingSize))),
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

	s.packetMtx.Lock()
	s.packetBuf.Write(packet)
	// logrus.Infof("pushPacket: %v len(buf): %d", packet, s.packetBuf.Len())
	s.packetMtx.Unlock()

	select {
	case s.packetWake <- struct{}{}:
	default:
	}
}

// Read data from the stream.
func (s *rawStream) Read(b []byte) (n int, err error) {
	for {
		s.packetMtx.Lock()
		n, err = s.packetBuf.Read(b)
		s.packetMtx.Unlock()
		if err == io.EOF {
			err = nil
		}
		if n != 0 || err != nil {
			return n, err
		}

		readDeadline := s.readDeadline
		var c <-chan time.Time
		if !readDeadline.IsZero() {
			now := time.Now()
			if readDeadline.Before(now) {
				return 0, context.DeadlineExceeded
			}

			fromNow := readDeadline.Sub(now)
			c = time.After(fromNow)
		}

		select {
		case <-s.ctx.Done():
			s.packetMtx.Lock()
			n, err = s.packetBuf.Read(b)
			s.packetMtx.Unlock()
			if n == 0 && err == nil {
				err = io.EOF
			}
			return
		case <-c:
			return 0, context.DeadlineExceeded
		case <-s.packetWake:
			// wakeup
		}
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
					pkt = xmitBuf.Get().([]byte)
					xpkt = pkt
					defer func() {
						xmitBuf.Put(pkt)
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
