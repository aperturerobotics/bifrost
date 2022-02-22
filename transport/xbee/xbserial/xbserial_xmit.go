package xbserial

import (
	"context"
	"encoding/binary"

	"github.com/pauleyj/gobee/api/rx"
	"github.com/pkg/errors"
)

// handleTxStatusFrame handles a tx status frame
func (x *XBeeSerial) handleTxStatusFrame(f *rx.TXStatus) {
	// drop frame, we don't care yet.
	// x.le.Debugf("tx status for packet %v retries %v delivery %v", f.ID(), f.Retries(), f.Delivery())
	select {
	case x.txStatusCh <- f:
	default:
	}
}

// XmitToAddr transmit to 64 bit address frame
type XmitToAddr struct {
	data []byte
}

// Bytes turn XmitToAddr frame into bytes
func (f *XmitToAddr) Bytes() ([]byte, error) {
	return f.data, nil
}

// TxToAddr transmits data to a 64 bit peer address.
//
// Use addr64 or addr16, not both.
// Special addr values:
// - 0x0 - Coordinator address
// - 0x000000000000FFFF - Broadcast address (64-bit only)
//
// https://www.digi.com/resources/documentation/Digidocs/90001942-13/reference/r_zigbee_frame_examples.htm
func (x *XBeeSerial) TxToAddr(
	ctx context.Context,
	addr64 uint64, addr16 uint16,
	// sourceEndpoint defaults to E8
	sourceEndpoint byte,
	// destinationEndpoint defaults to E8
	destinationEndpoint byte,
	// clusterID defaults to 0x11
	clusterID uint16,
	// profileID defaults to C1 05
	profileID uint16,
	// broadcastRadius defaults to 00
	// limits the maximum number of retransmission hops.
	broadcastRadius byte,
	// options defaults to 00 (none)
	options byte,
	// data is the packet to transmit to the peer
	data []byte,
) error {
	f := &XmitToAddr{}
	f.data = make([]byte, len(data)+1+1+8+2+1+1+2+2+1+1)
	// write frame type (10) (1 byte) [0]
	f.data[0] = 0x11
	// write frame ID (0) (1 byte) [1]
	frameID := data[0]
	if frameID == 0 {
		frameID = 0x23
	}
	f.data[1] = frameID
	// write 64 bit destination address (8 bytes)
	binary.BigEndian.PutUint64(f.data[2:], addr64)
	// write 16 bit destination address (2)
	if addr16 == 0 {
		f.data[10] = 0xFF
		f.data[11] = 0xFE
	} else {
		binary.BigEndian.PutUint16(f.data[10:], addr16)
	}
	// write source endpoint byte
	if sourceEndpoint == 0 {
		f.data[12] = 0xE8
	} else {
		f.data[12] = sourceEndpoint
	}
	// write destination endpoint byte
	if destinationEndpoint == 0 {
		f.data[13] = 0xE8
	} else {
		f.data[13] = destinationEndpoint
	}
	// write cluster ID
	if clusterID == 0 {
		f.data[14] = 0
		f.data[15] = 0x11
	} else {
		binary.LittleEndian.PutUint16(f.data[14:], clusterID)
	}
	// write profile id
	if profileID == 0 {
		f.data[16] = 0xC1
		f.data[17] = 0x05
	}
	// write broadcast radius (max hops) (1 byte)
	if broadcastRadius >= 0 {
		f.data[18] = broadcastRadius
	}
	// write options (1) (1 byte)
	// 0x01: no retries
	// 0x20: enable aps encrypt (if EE = 1)
	// 0x40: use extended transmission timeout
	f.data[19] = options
	// write data len(data)
	copy(f.data[20:], data)

	// x.le.Debugf("writing tx frame datalen(%d) len(%d): %x", len(data), len(f.data), f.data)
	x.txMtx.Lock()
	defer x.txMtx.Unlock()

	_, err := x.WriteFrame(f)
	if err != nil {
		return err
	}

WaitTxLoop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case statusPkt := <-x.txStatusCh:
			if statusPkt.ID() == frameID {
				// x.le.Debugf("frame id %v status %v", frameID, statusPkt.Delivery())
				if statusPkt.Delivery() != 0 {
					err := errors.Errorf("delivery failed with status %v", statusPkt.Delivery())
					x.le.WithError(err).Warn("error txing packet")
					return err
				}

				break WaitTxLoop
			}
		}
	}

	return nil
}
