package stream_packet

import (
	"context"
	"encoding/binary"
	"io"
	"sync"

	"github.com/aperturerobotics/bifrost/stream"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// Session wraps a stream in a session.
type Session struct {
	stream.Stream
	ctx            context.Context
	ctxCancel      context.CancelFunc
	sendMtx        sync.Mutex
	readMtx        sync.Mutex
	maxMessageSize uint32
}

// NewSession builds a new session.
func NewSession(
	stream stream.Stream,
	maxMessageSize uint32,
) *Session {
	return &Session{
		Stream:         stream,
		maxMessageSize: maxMessageSize,
	}
}

// SendMsg tries to send a message on the wire.
func (s *Session) SendMsg(msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	pktBuf := make([]byte, len(data)+4)
	copy(pktBuf[4:], data)
	msgLen := uint32(len(data))
	binary.LittleEndian.PutUint32(pktBuf, msgLen)

	s.sendMtx.Lock()
	defer s.sendMtx.Unlock()

	if _, err := s.Stream.Write(pktBuf); err != nil {
		return err
	}
	return nil
}

// RecvMsg tries to receive a message on the wire.
func (s *Session) RecvMsg(msg proto.Message) error {
	data := make([]byte, 4)
	s.readMtx.Lock()
	defer s.readMtx.Unlock()

	if _, err := io.ReadFull(s.Stream, data); err != nil {
		return err
	}

	messageLen := binary.LittleEndian.Uint32(data)
	if messageLen > 0 {
		if messageLen > s.maxMessageSize {
			return errors.Errorf("invalid message len: %d", messageLen)
		}

		data = make([]byte, messageLen)
		if _, err := io.ReadFull(s.Stream, data); err != nil {
			return err
		}

		return proto.Unmarshal(data, msg)
	}

	msg.Reset()
	return nil
}
