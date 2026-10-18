package wasivm_plugin

import (
	"bytes"
	"context"
	"testing"

	"github.com/aperturerobotics/starpc/srpc"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// frameSource serves the Display method as the WasiVM plugin does, but sends
// pre-encoded frame bytes, so a test controls exactly what is on the wire.
type frameSource struct {
	frame []byte
}

// InvokeMethod sends the frame once and ends the stream.
func (s *frameSource) InvokeMethod(serviceID, methodID string, strm srpc.Stream) (bool, error) {
	if serviceID != s4wave_vm.SRPCWasiVmDisplayServiceServiceID || methodID != "Display" {
		return false, nil
	}
	return true, strm.MsgSend(srpc.NewRawMessage(s.frame, true))
}

// TestDisplayFrameRelayedUnchanged checks a hop that forwards the Display RPC
// with srpc.NewClientInvoker, as the core plugin does for a plugin's
// resource, delivers a frame's bytes as the plugin encoded them. The frame
// is written with its fields out of order, which decoding and encoding the
// frame would put back in order.
func TestDisplayFrameRelayedUnchanged(t *testing.T) {
	// Encode a frame the way the plugin does, then put the screen height first.
	frame := &s4wave_vm.WasiVmDisplayFrame{
		ScreenWidth:  640,
		ScreenHeight: 480,
		Copies:       []*s4wave_vm.WasiVmDisplayCopy{{X: 1, Y: 2, Width: 3, Height: 4, ToX: 5, ToY: 6}},
		Regions: []*s4wave_vm.WasiVmDisplayRegion{{
			X: 10, Y: 20, Width: 256, Height: 256,
			Rgba: bytes.Repeat([]byte{1, 2, 3, 255}, 256*256),
		}},
	}
	encoded, err := frame.MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	wire := append([]byte{0x10, 0xe0, 0x03}, encoded...)

	// Control: decoding and encoding the frame puts the fields back in order.
	var decoded s4wave_vm.WasiVmDisplayFrame
	if err := decoded.UnmarshalVT(wire); err != nil {
		t.Fatal(err)
	}
	if again, err := decoded.MarshalVT(); err != nil || bytes.Equal(again, wire) {
		t.Fatalf("the field order survived a decode and encode (err %v)", err)
	}

	// Relay the plugin's Display RPC through a client invoker.
	pluginClient := srpc.NewClient(srpc.NewServerPipe(srpc.NewServer(&frameSource{frame: wire})))
	relayClient := srpc.NewClient(srpc.NewServerPipe(srpc.NewServer(srpc.NewClientInvoker(pluginClient))))

	// The viewer receives the plugin's bytes without a decode on the way.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	strm, err := relayClient.NewStream(ctx, s4wave_vm.SRPCWasiVmDisplayServiceServiceID, "Display", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer strm.Close()
	got := srpc.NewRawMessage(nil, true)
	if err := strm.MsgRecv(got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.GetData(), wire) {
		t.Fatalf("the relayed frame is %d bytes, want the plugin's %d bytes unchanged", len(got.GetData()), len(wire))
	}
}
