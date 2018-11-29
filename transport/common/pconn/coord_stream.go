package pconn

import (
	"encoding/binary"
	"io"

	"github.com/aperturerobotics/bifrost/stream"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// maxCoordStreamPlen is the maximum length of a coordination stream packet.
// expect at most 2 uint32
// 2000 bytes to be safe
var maxCoordStreamPlen = uint32(2000)

// maxInflightEstablishes is the maximum number of in-flight stream establishes
var maxInflightEstablishes = uint32(5)

// coordinationStreamPump manages the coordination stream.
func (l *Link) coordinationStreamPump(coordStrm stream.Stream) {
	defer l.Close()

	l.rawStreamsMtx.Lock()
	l.coordStream = coordStrm
	l.rawStreamsMtx.Unlock()

	plenBuf := make([]byte, 4)
	var pkt CoordinationStreamPacket
	var pktBuf []byte

	for {
		// Coordination stream is stream-oriented, not packet oriented.
		// Read uint32 packet length header.
		if _, err := io.ReadFull(coordStrm, plenBuf); err != nil {
			// l.le.WithError(err).Warn("coordination stream: cannot read packet len")
			return
		}

		// Decode packet length
		packetLen := binary.LittleEndian.Uint32(plenBuf)
		if packetLen > maxCoordStreamPlen {
			l.le.Warnf(
				"coordination stream: packet of length %d > %d (max)",
				packetLen,
				maxCoordStreamPlen,
			)
			return
		}

		// Packet buffer extend or slice
		if cap(pktBuf) < int(packetLen) {
			pktBuf = make([]byte, packetLen)
		} else {
			pktBuf = pktBuf[:packetLen]
		}

		// Read packet body
		if _, err := io.ReadFull(coordStrm, pktBuf); err != nil {
			l.le.
				WithError(err).
				WithField("packet-length", packetLen).
				Warn("coordination stream: cannot read packet")
			return
		}

		// Decode packet body
		pkt.Reset()
		if err := proto.Unmarshal(pktBuf, &pkt); err != nil {
			l.le.
				WithError(err).
				Warn("coordination stream: cannot unmarshal packet")
			return
		}

		var err error
		pktType := pkt.GetPacketType()
		// l.le.Debugf("coordination stream: read packet of length %d type %s", packetLen, pktType.String())
		switch pktType {
		case CoordPacketType_CoordPacketType_RSTREAM_ESTABLISH:
			err = l.handleCoordRawStreamEstablish(pkt.GetRawStreamEstablish())
		case CoordPacketType_CoordPacketType_RSTREAM_ACK:
			err = l.handleCoordRawStreamAck(pkt.GetRawStreamAck())
		case CoordPacketType_CoordPacketType_RSTREAM_CLOSE:
			err = l.handleCoordRawStreamClose(pkt.GetRawStreamClose())
		default:
			l.le.
				WithField("packet-type", pktType.String()).
				Warn("coordination stream: unknown packet type dropped")
			continue
		}

		if err != nil {
			l.le.
				WithError(err).
				Warn("error handling coordination message")
		}
	}
}

// handleCoordRawStreamEstablish handles a raw stream establish message.
// CoordPacketType_CoordPacketType_RSTREAM_ESTABLISH
func (l *Link) handleCoordRawStreamEstablish(pkt *RawStreamEstablish) error {
	l.rawStreamsMtx.Lock()
	defer l.rawStreamsMtx.Unlock()

	// check the length of the accept queue
	if len(l.rawStreamEstablishQueueInc) >= int(maxInflightEstablishes) {
		return errors.New("maximum in-flight stream establishes exceeded")
	}

	initStreamID := pkt.GetInitiatorStreamId()
	// pick the next stream ID to use
	localStreamID := l.nextRawStreamID
	l.nextRawStreamID++

	// register next stream
	rstrm := l.constructRawStream(localStreamID, nil)
	rstrm.SetRemoteStreamID(initStreamID)
	l.rawStreams[localStreamID] = rstrm

	l.rawStreamEstablishQueueInc = append(l.rawStreamEstablishQueueInc, rstrm)

	// notify the acceptor that the queue is filled
	select {
	case l.acceptStreamCh <- nil:
	default:
	}

	return nil
}

// constructRawStream constructs the rawStream object with a local stream ID
func (l *Link) constructRawStream(localStreamID uint32, establishCb func(err error)) *rawStream {
	return newRawStream(
		l.ctx,
		localStreamID,
		l.mtu,
		establishCb,
		func(data []byte) error {
			data = append(data, byte(PacketType_PacketType_RAW))
			_, err := l.writer(data, l.addr)
			return err
		}, func(rs *rawStream) {
			l.rawStreamsMtx.Lock()
			if !rs.closed {
				if l.rawStreams != nil {
					if est := l.rawStreams[rs.localStreamID]; est == rs {
						l.le.WithField("local-stream-id", localStreamID).Debug("raw stream closed")
						_ = l.writeStreamClosePacket(rs.remoteStreamID)
						delete(l.rawStreams, rs.localStreamID)
					}
				}
				rs.markClosed()
			}
			l.rawStreamsMtx.Unlock()
		},
	)
}

// Validate validates the raw stream ack packet.
func (p *RawStreamAck) Validate() error {
	if p.GetInitiatorStreamId() == 0 {
		return errors.New("ack: initiator stream id cannot be empty")
	}
	if p.GetAckStreamId() == 0 {
		if p.GetAckError() == "" {
			return errors.New("ack: ack stream id cannot be empty without an error")
		}
	}

	return nil
}

// handleCoordRawStreamEstablish handles a raw stream establish ack message.
// CoordPacketType_CoordPacketType_RSTREAM_ACK
func (l *Link) handleCoordRawStreamAck(pkt *RawStreamAck) error {
	localStreamID := pkt.GetInitiatorStreamId()
	if err := pkt.Validate(); err != nil {
		return err
	}

	l.rawStreamsMtx.Lock()
	defer l.rawStreamsMtx.Unlock()

	rstrm, rstrmOk := l.rawStreams[localStreamID]
	if !rstrmOk {
		if pkt.GetAckError() != "" {
			return nil
		}

		return errors.Errorf("ack: unknown local stream %d", localStreamID)
	}

	// already ack'd
	if rstrm.remoteStreamID != 0 {
		return nil
	}

	var estErr error
	if pkt.GetAckError() != "" {
		rstrm.markClosed()
		delete(l.rawStreams, localStreamID)
		estErr = errors.Errorf("remote establish error: %s", pkt.GetAckError())
	} else {
		rstrm.SetRemoteStreamID(pkt.GetAckStreamId())
	}
	if rstrm.establishCb != nil {
		go rstrm.establishCb(estErr)
	}
	if l.inflightRawStreamEstablishOut != 0 {
		l.inflightRawStreamEstablishOut--
	}

	for l.inflightRawStreamEstablishOut < maxInflightEstablishes {
		if len(l.rawStreamEstablishQueueOut) == 0 {
			break
		}

		o := l.rawStreamEstablishQueueOut[len(l.rawStreamEstablishQueueOut)-1]
		l.rawStreamEstablishQueueOut[len(l.rawStreamEstablishQueueOut)-1] = nil
		l.rawStreamEstablishQueueOut = l.rawStreamEstablishQueueOut[:len(l.rawStreamEstablishQueueOut)-1]
		if err := l.writeRawStreamEstablish(o); err != nil {
			l.le.WithError(err).Warn("error writing raw stream establish")
		} else {
			l.inflightRawStreamEstablishOut++
		}
	}

	return estErr
}

// writeStreamClosePacket writes the stream close packet to the coordination stream.
func (l *Link) writeStreamClosePacket(remoteStreamID uint32) error {
	return l.writeCoordStreamPacket(&CoordinationStreamPacket{
		PacketType: CoordPacketType_CoordPacketType_RSTREAM_CLOSE,
		RawStreamClose: &RawStreamClose{
			StreamId: remoteStreamID,
		},
	})
}

// writeCoordStreamPacket writes a packet to the coord stream
// assumes rawStreamsMtx is locked
func (l *Link) writeCoordStreamPacket(pkt *CoordinationStreamPacket) error {
	if l.coordStream == nil {
		return nil
	}

	// encode packet
	pktDat, err := proto.Marshal(pkt)
	if err != nil {
		return err
	}

	// encode packet len
	pktBuf := make([]byte, 4+len(pktDat))
	binary.LittleEndian.PutUint32(pktBuf[:4], uint32(len(pktDat)))
	copy(pktBuf[4:], pktDat)
	if _, err := l.coordStream.Write(pktBuf); err != nil {
		return err
	}

	return nil
}

// handleCoordRawStreamClose handles a raw stream close message.
// CoordPacketType_CoordPacketType_RSTREAM_CLOSE
func (l *Link) handleCoordRawStreamClose(pkt *RawStreamClose) error {
	l.rawStreamsMtx.Lock()
	defer l.rawStreamsMtx.Unlock()

	localStreamID := pkt.GetStreamId()
	localStream, localStreamOK := l.rawStreams[localStreamID]
	if !localStreamOK {
		return nil
	}

	_ = localStream.Close()
	if closeErr := pkt.GetCloseError(); closeErr != "" {
		l.le.WithError(errors.New(closeErr)).Warn("raw stream closed with error")
		return nil
	}
	return nil
}

// openRawStream implements OpenStream for a raw stream.
func (l *Link) openRawStream() (stream.Stream, error) {
	l.rawStreamsMtx.Lock()

	// pick the next stream ID to use
	localStreamID := l.nextRawStreamID
	l.nextRawStreamID++

	// register next stream
	errCh := make(chan error, 1)
	rstrm := l.constructRawStream(localStreamID, func(err error) {
		// err is the link construction error
		errCh <- err
	})
	l.rawStreams[localStreamID] = rstrm

	// transmit raw stream establish
	if l.coordStream == nil ||
		l.inflightRawStreamEstablishOut >= maxInflightEstablishes {
		l.rawStreamEstablishQueueOut = append(l.rawStreamEstablishQueueOut, rstrm)
	} else {
		l.inflightRawStreamEstablishOut++
		if err := l.writeRawStreamEstablish(rstrm); err != nil {
			l.rawStreamsMtx.Unlock()
			return nil, err
		}
	}

	// wait for the result
	l.rawStreamsMtx.Unlock()

	select {
	case <-l.ctx.Done():
		go rstrm.Close()
		return nil, l.ctx.Err()
	case err := <-errCh:
		if err == nil {
			return rstrm, nil
		}

		return nil, err
	}
}

func (l *Link) writeRawStreamEstablish(rs *rawStream) error {
	err := l.writeCoordStreamPacket(&CoordinationStreamPacket{
		PacketType: CoordPacketType_CoordPacketType_RSTREAM_ESTABLISH,
		RawStreamEstablish: &RawStreamEstablish{
			InitiatorStreamId: rs.localStreamID,
		},
	})
	if err != nil {
		rs.markClosed()
		delete(l.rawStreams, rs.localStreamID)
		return err
	}
	return nil
}

func (l *Link) writeRawStreamAck(rs *rawStream) error {
	err := l.writeCoordStreamPacket(&CoordinationStreamPacket{
		PacketType: CoordPacketType_CoordPacketType_RSTREAM_ACK,
		RawStreamAck: &RawStreamAck{
			InitiatorStreamId: rs.remoteStreamID,
			AckStreamId:       rs.localStreamID,
		},
	})
	if err != nil {
		rs.markClosed()
		delete(l.rawStreams, rs.localStreamID)
		return err
	}
	return nil
}

// drainIncomingEstablishQueue attempts to pop one value off the establish queue.
func (l *Link) drainIncomingEstablishQueue() *rawStream {
	if len(l.rawStreamEstablishQueueInc) == 0 {
		return nil
	}

	strm := l.rawStreamEstablishQueueInc[0]
	werr := l.writeRawStreamAck(strm)
	// shift
	for i := 0; i < len(l.rawStreamEstablishQueueInc)-1; i++ {
		l.rawStreamEstablishQueueInc[i] = l.rawStreamEstablishQueueInc[i+1]
	}
	l.rawStreamEstablishQueueInc[len(l.rawStreamEstablishQueueInc)-1] = nil
	l.rawStreamEstablishQueueInc = l.rawStreamEstablishQueueInc[:len(l.rawStreamEstablishQueueInc)-1]

	if werr != nil {
		l.le.WithError(werr).Warn("error accepting raw stream")
		return nil
	}
	return strm
}
