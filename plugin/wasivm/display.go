package wasivm_plugin

import (
	"context"

	"github.com/s4wave/spacewave/plugin/wasivm/display"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// vmDisplay serves the X display of one VM on each of its resources. A viewer
// that attaches before the VM runs, or before its X server listens, waits for
// both.
type vmDisplay struct {
	consoles  *consoleTable
	objectKey string
}

// Display connects to X display 0 of the running VM, streams its screen to
// the viewer and injects the viewer's input, until the X server closes the
// connection or the viewer leaves.
func (v *vmDisplay) Display(strm s4wave_vm.SRPCWasiVmDisplayService_DisplayStream) error {
	// Wait for the VM to run.
	ctx, cancel := context.WithCancel(strm.Context())
	defer cancel()
	c, err := v.consoles.wait(ctx, v.objectKey)
	if err != nil {
		return err
	}

	// Connect to its X server once the server listens.
	conn, err := c.kernel.DialUnix(ctx, display.SocketPath)
	if err != nil {
		return err
	}
	client, err := display.New(conn)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()

	// Inject the viewer's input until it leaves.
	go func() {
		defer cancel()
		for {
			in, err := strm.Recv()
			if err != nil {
				return
			}
			client.Input(in)
		}
	}()

	// Stream the screen.
	return client.Run(ctx, strm.Send)
}

// _ is a type assertion.
var _ s4wave_vm.SRPCWasiVmDisplayServiceServer = (*vmDisplay)(nil)
