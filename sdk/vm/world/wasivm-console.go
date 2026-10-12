package s4wave_vm_world

import (
	"context"
	"io"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/aperturerobotics/util/broadcast"
	bifrost_rpc "github.com/s4wave/spacewave/net/rpc"
	s4wave_terminal "github.com/s4wave/spacewave/sdk/terminal"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// wasiVmScrollback is the output a console keeps at least for viewers that
// attach after it was written.
const wasiVmScrollback = 64 * 1024

// wasiVmConsole is the terminal of one WasiVM run. It keeps the recent output
// for viewers that attach later and fans new output and the exit out to every
// attached viewer.
type wasiVmConsole struct {
	// input feeds the kernel's terminal.
	input io.Writer
	// resize sets the size of the kernel's terminal.
	resize func(cols, rows uint32)

	// bcast guards the fields below and wakes viewers when they change.
	bcast broadcast.Broadcast
	// out holds the most recent output, which ends at offset end of the
	// output stream. Bytes within len(out) are never rewritten, so a viewer
	// may send a subslice after releasing the lock.
	out []byte
	// end is the number of bytes written since the run began.
	end uint64
	// exited is set once the kernel has exited.
	exited bool
	// exitCode is the exit code of init.
	exitCode int32
	// exitError describes a run that failed instead of exiting.
	exitError string
}

// Write appends kernel output to the console.
func (c *wasiVmConsole) Write(p []byte) (int, error) {
	c.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		c.out = append(c.out, p...)
		c.end += uint64(len(p))
		if len(c.out) > 2*wasiVmScrollback {
			c.out = append([]byte(nil), c.out[len(c.out)-wasiVmScrollback:]...)
		}
		broadcast()
	})
	return len(p), nil
}

// exit records how the run ended and ends every viewer stream. runErr is the
// error that ended a failed run.
func (c *wasiVmConsole) exit(code int32, runErr error) {
	c.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		c.exited = true
		c.exitCode = code
		if runErr != nil {
			c.exitError = runErr.Error()
		}
		broadcast()
	})
}

// Console attaches a viewer. It replays the retained output, then streams new
// output until the VM exits or the viewer closes. Input and size changes from
// the viewer go to the kernel's terminal.
func (c *wasiVmConsole) Console(strm s4wave_vm.SRPCWasiVmConsoleService_ConsoleStream) error {
	// Tell the viewer the console is ready.
	ctx, cancel := context.WithCancel(strm.Context())
	defer cancel()
	if err := strm.Send(&s4wave_terminal.TerminalFrame{
		Kind: s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_READY,
	}); err != nil {
		return err
	}

	// Forward viewer input and size until the viewer closes.
	go func() {
		defer cancel()
		for {
			frame, err := strm.Recv()
			if err != nil {
				return
			}
			switch frame.GetKind() {
			case s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_INPUT:
				if _, err := c.input.Write(frame.GetData()); err != nil {
					return
				}
			case s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_OPEN,
				s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_RESIZE:
				if frame.GetCols() != 0 && frame.GetRows() != 0 {
					c.resize(frame.GetCols(), frame.GetRows())
				}
			case s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_CLOSE:
				return
			}
		}
	}()

	// Send output from the oldest retained byte until the VM exits.
	var next uint64
	for {
		var data []byte
		var exit *s4wave_terminal.TerminalFrame
		var waitCh <-chan struct{}
		c.bcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			// Take the output the viewer has not seen, and the exit once
			// the VM has exited.
			start := c.end - uint64(len(c.out))
			next = max(next, start)
			data = c.out[next-start:]
			next = c.end
			if c.exited {
				exit = &s4wave_terminal.TerminalFrame{
					Kind:     s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_EXIT,
					ExitCode: c.exitCode,
					Error:    c.exitError,
				}
			}
			waitCh = getWaitCh()
		})
		if len(data) != 0 {
			if err := strm.Send(&s4wave_terminal.TerminalFrame{
				Kind: s4wave_terminal.TerminalFrameKind_TERMINAL_FRAME_KIND_OUTPUT,
				Data: data,
			}); err != nil {
				return err
			}
		}
		if exit != nil {
			return strm.Send(exit)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-waitCh:
		}
	}
}

// wasiVmConsoleForwarder serves the console on a viewer's object resource by
// forwarding each call to the route of the VM's running instance. A call made
// while the VM is not running waits for its console to appear.
type wasiVmConsoleForwarder struct {
	inv       srpc.Invoker
	serviceID string
}

// newWasiVmConsoleForwarder forwards consoles to the VM process serving
// under route on the bus b.
func newWasiVmConsoleForwarder(b bus.Bus, route string) *wasiVmConsoleForwarder {
	return &wasiVmConsoleForwarder{
		inv:       bifrost_rpc.NewInvoker(b, "", true),
		serviceID: route + s4wave_vm.SRPCWasiVmConsoleServiceServiceID,
	}
}

// Console forwards the viewer stream to the running VM's console.
func (f *wasiVmConsoleForwarder) Console(strm s4wave_vm.SRPCWasiVmConsoleService_ConsoleStream) error {
	found, err := f.inv.InvokeMethod(f.serviceID, "Console", strm)
	if err == nil && !found {
		return srpc.ErrUnimplemented
	}
	return err
}

// _ is a type assertion
var (
	_ s4wave_vm.SRPCWasiVmConsoleServiceServer = (*wasiVmConsole)(nil)
	_ s4wave_vm.SRPCWasiVmConsoleServiceServer = (*wasiVmConsoleForwarder)(nil)
)
