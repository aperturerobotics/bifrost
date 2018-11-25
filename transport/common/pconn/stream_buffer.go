package pconn

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/djherbis/buffer"
)

// keep 6000 packets
// 1 packet = 1500 byte maximum
// 6000*1500 ~= 10mb packet ring
var packetRingSize = 6000 * 1500

// PacketBuffer implements a net.Conn backed by a packet buffer
type PacketBuffer struct {
	ctx context.Context

	packetMtx  sync.Mutex
	packetBuf  buffer.Buffer
	packetWake chan struct{}

	readDeadline, writeDeadline time.Time
	writeFn                     func(data []byte) error

	// remoteStreamID is the remote stream id
	remoteStreamID uint32

	// headerVarint is the header varint
	// the bytes are reversed.
	// empty until remote stream ID is known
	headerVarint []byte
}

func NewPacketBuffer(
	ctx context.Context,
	writeFn func(data []byte) error,
) *PacketBuffer {
	return &PacketBuffer{
		ctx:        ctx,
		writeFn:    writeFn,
		packetWake: make(chan struct{}, 1),
		packetBuf:  buffer.NewRing(buffer.New(int64(packetRingSize))),
	}
}

// PushPacket pushes a packet to the stream, dropping the oldest packet if necessary.
func (s *PacketBuffer) PushPacket(packet []byte) {
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
func (s *PacketBuffer) Read(b []byte) (n int, err error) {
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
func (s *PacketBuffer) Write(b []byte) (n int, err error) {
	blen := len(b)

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
func (s *PacketBuffer) SetReadDeadline(t time.Time) error {
	s.readDeadline = t
	return nil
}

// SetWriteDeadline sets the write deadline as defined by
// A zero time value disables the deadline.
func (s *PacketBuffer) SetWriteDeadline(t time.Time) error {
	s.writeDeadline = t
	return nil
}

// SetDeadline sets both read and write deadlines as defined by
// A zero time value disables the deadlines.
func (s *PacketBuffer) SetDeadline(t time.Time) error {
	s.readDeadline = t
	s.writeDeadline = t
	return nil
}

// SetRemoteStreamID sets the remote stream ID varint trailer
func (s *PacketBuffer) SetRemoteStreamID(id uint32) {
	s.remoteStreamID = id
	s.headerVarint = encodeRawStreamIDVarint(id)
}
