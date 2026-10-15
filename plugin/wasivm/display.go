package wasivm_plugin

import (
	"context"

	"github.com/s4wave/spacewave/plugin/wasivm/display"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// vmDisplay serves the display of one VM on each of its resources. A viewer
// that attaches before the VM runs, or before the guest brings up a display,
// waits for both.
type vmDisplay struct {
	consoles  *consoleTable
	objectKey string
}

// screen is a display a viewer attaches to: X display 0 or the Wayland
// compositor.
type screen interface {
	// Input injects a pointer or key event.
	Input(in *s4wave_vm.WasiVmDisplayInput)
	// Run streams the whole screen, then each area that changes, until the
	// display or ctx ends.
	Run(ctx context.Context, send func(*s4wave_vm.WasiVmDisplayFrame) error) error
}

// attached is a screen with the function that detaches from it, or the error
// that kept the viewer from attaching.
type attached struct {
	screen  screen
	release func()
	err     error
}

// Display attaches to the display of the running VM, streams its screen to
// the viewer and injects the viewer's input, until the display ends or the
// viewer leaves.
func (v *vmDisplay) Display(strm s4wave_vm.SRPCWasiVmDisplayService_DisplayStream) error {
	// Wait for the VM to run and bring up a display.
	ctx, cancel := context.WithCancel(strm.Context())
	defer cancel()
	c, err := v.consoles.wait(ctx, v.objectKey)
	if err != nil {
		return err
	}
	a := c.attachScreen(ctx)
	if a.err != nil {
		return a.err
	}
	defer a.release()

	// Inject the viewer's input until it leaves.
	go func() {
		defer cancel()
		for {
			in, err := strm.Recv()
			if err != nil {
				return
			}
			a.screen.Input(in)
		}
	}()

	// Stream the screen.
	return a.screen.Run(ctx, strm.Send)
}

// attachScreen waits for the first display the guest brings up: X display 0
// once its server listens, or the Wayland compositor once a client connects.
// It detaches from the other.
func (c *console) attachScreen(ctx context.Context) attached {
	// Wait for each display.
	waitCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan attached, 2)
	waits := 1
	go func() { results <- c.attachX(waitCtx) }()
	if comp := c.wayland; comp != nil {
		waits++
		go func() {
			results <- attached{screen: comp, release: func() {}, err: comp.WaitClient(waitCtx)}
		}()
	}

	// Keep the first, then detach from the other once its wait ends.
	first := <-results
	cancel()
	for range waits - 1 {
		if late := <-results; late.err == nil {
			late.release()
		}
	}
	return first
}

// attachX connects to X display 0 once its server listens.
func (c *console) attachX(ctx context.Context) attached {
	// Connect, then greet the server as an X client.
	conn, err := c.kernel.DialUnix(ctx, display.SocketPath)
	if err != nil {
		return attached{err: err}
	}
	client, err := display.New(conn)
	if err != nil {
		_ = conn.Close()
		return attached{err: err}
	}
	return attached{screen: client, release: client.Close}
}

// _ is a type assertion.
var _ s4wave_vm.SRPCWasiVmDisplayServiceServer = (*vmDisplay)(nil)
