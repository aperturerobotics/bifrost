package simulate

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/stream"
	stream_echo "github.com/aperturerobotics/bifrost/stream/echo"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/pkg/errors"
)

// TestConnectivity tests for basic connectivity between two peers.
// Uses an echo test: dials the Echo controller.
func TestConnectivity(ctx context.Context, px0, px1 *Peer) error {
	tb0 := px0.testbed

	msv1, ms1Ref, err := bus.ExecOneOff(
		ctx,
		tb0.Bus,
		link.NewOpenStreamWithPeer(
			stream_echo.DefaultProtocolID,
			px0.GetPeerID(),
			px1.GetPeerID(),
			0,
			stream.OpenOpts{Reliable: true, Encrypted: true},
		),
		false,
		nil,
	)
	if err != nil {
		return err
	}
	defer ms1Ref.Release()

	mns1 := msv1.GetValue().(link.MountedStream)
	ms1 := mns1.GetStream()
	// expect px0 stream remote peer to equal px1
	mns1rp := mns1.GetLink().GetRemotePeer().Pretty()
	if px1p := px1.GetPeerID().Pretty(); px1p != mns1rp {
		return errors.Errorf(
			"stream on p0 remote peer id %s != expected %s",
			mns1rp,
			px1p,
		)
	}
	// expect px0 stream local peer to equal px0
	mns1lp := mns1.GetLink().GetLocalPeer().Pretty()
	if px0p := px0.GetPeerID().Pretty(); px0p != mns1lp {
		return errors.Errorf(
			"stream on p0 local peer id %s != expected %s",
			mns1lp,
			px0p,
		)
	}

	data := []byte("testing 1234")
	_, err = ms1.Write(data)
	if err != nil {
		return err
	}

	// expect remote to echo back exactly len(data) bytes
	outData := make([]byte, len(data)*2)
	on, oe := ms1.Read(outData)
	if oe != nil {
		return oe
	}

	outData = outData[:on]
	if on != len(data) {
		return errors.Errorf(
			"length incorrect received %v != %v data: %v",
			on,
			len(data),
			string(outData),
		)
	}
	return nil
}
