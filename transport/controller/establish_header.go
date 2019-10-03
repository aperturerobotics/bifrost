package transport_controller

import (
	"io"

	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// NewStreamEstablish constructs a new StreamEstablish message.
func NewStreamEstablish(protocolID protocol.ID) *StreamEstablish {
	return &StreamEstablish{ProtocolId: string(protocolID)}
}

func writeStreamEstablishHeader(w io.Writer, msg *StreamEstablish) (int, error) {
	dat, err := proto.Marshal(msg)
	if err != nil {
		return 0, err
	}

	lenVarInt := proto.EncodeVarint(uint64(len(dat)))
	outBuf := make([]byte, len(dat)+len(lenVarInt))
	copy(outBuf, lenVarInt)
	copy(outBuf[len(lenVarInt):], dat)

	return w.Write(outBuf)
}

func readAtLeast(r io.Reader, n, min int, buf []byte) (int, error) {
	for n < min {
		nr, err := r.Read(buf[n:])
		if err != nil {
			return n, err
		}
		n += nr
	}
	return n, nil
}

func readStreamEstablishHeader(r io.Reader) (*StreamEstablish, error) {
	b := make([]byte, 4)
	var err error
	_, err = readAtLeast(r, 0, 4, b)
	if err != nil {
		return nil, err
	}

	// Read the header length varint
	headerLen, headerLenBytes := proto.DecodeVarint(b)
	if headerLenBytes == 0 {
		return nil, err
	}
	if headerLenBytes > len(b) { // this should not be possible
		headerLenBytes = len(b)
	}

	// header len is at most 100,000 bytes
	if int(headerLen) > streamEstablishMaxPacketSize || headerLen == 0 {
		return nil, errors.Errorf(
			"stream establish header length invalid: %d (expected <= %d)",
			headerLen,
			streamEstablishMaxPacketSize,
		)
	}

	headerBuf := make([]byte, int(headerLen))
	copy(headerBuf, b[headerLenBytes:])
	n := len(b) - headerLenBytes
	if _, err := readAtLeast(r, n, int(headerLen), headerBuf); err != nil {
		return nil, err
	}

	// logrus.Infof("read stream establish: %v", headerBuf)
	// decode stream establish header
	estHeader := &StreamEstablish{}
	if err := proto.Unmarshal(headerBuf, estHeader); err != nil {
		return nil, err
	}

	return estHeader, nil
}
