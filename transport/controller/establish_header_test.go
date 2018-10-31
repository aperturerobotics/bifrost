package transport_controller

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"
)

// TestReadStreamEstablishHeader tests reading a stream establish header.
func TestReadStreamEstablishHeader(t *testing.T) {
	buf := &bytes.Buffer{}
	obj := &StreamEstablish{ProtocolId: "testing"}
	dat, err := proto.Marshal(obj)
	if err != nil {
		t.Fatal(err.Error())
	}

	buf.Write(proto.EncodeVarint(uint64(len(dat))))
	buf.Write(dat)
	readEst, err := readStreamEstablishHeader(buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	if readEst.ProtocolId != obj.ProtocolId {
		t.Fatalf("object decoded incorrectly: %s != %s", readEst.ProtocolId, obj.ProtocolId)
	}
}
