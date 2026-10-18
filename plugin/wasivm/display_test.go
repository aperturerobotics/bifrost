package wasivm_plugin

import (
	"context"
	"testing"
	"time"

	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// testStream is a viewer's end of a display stream: it records the frames
// sent and receives the inputs pushed to in.
type testStream struct {
	s4wave_vm.SRPCWasiVmDisplayService_DisplayStream
	ctx    context.Context
	frames chan *s4wave_vm.WasiVmDisplayFrame
	in     chan *s4wave_vm.WasiVmDisplayInput
}

// Context returns the stream's context.
func (s *testStream) Context() context.Context { return s.ctx }

// Send records a frame.
func (s *testStream) Send(f *s4wave_vm.WasiVmDisplayFrame) error {
	s.frames <- f
	return nil
}

// Recv returns the next pushed input, or the context's error.
func (s *testStream) Recv() (*s4wave_vm.WasiVmDisplayInput, error) {
	select {
	case in := <-s.in:
		return in, nil
	case <-s.ctx.Done():
		return nil, context.Cause(s.ctx)
	}
}

// testScreen sends n frames, one after another, and records the input.
type testScreen struct {
	n     int
	input chan *s4wave_vm.WasiVmDisplayInput
}

// Input records an event.
func (s *testScreen) Input(in *s4wave_vm.WasiVmDisplayInput) { s.input <- in }

// Run sends n frames of increasing width.
func (s *testScreen) Run(_ context.Context, send func(*s4wave_vm.WasiVmDisplayFrame) error) error {
	for i := range s.n {
		if err := send(&s4wave_vm.WasiVmDisplayFrame{ScreenWidth: uint32(i + 1)}); err != nil { //nolint:gosec // small.
			return err
		}
	}
	return nil
}

// TestServeScreenPaced checks the viewer gets frames up to framesAhead ahead
// of its reports of drawing them, while its input passes through at once.
func TestServeScreenPaced(t *testing.T) {
	// Serve a screen with four frames.
	const n = 4
	strm := &testStream{
		ctx:    t.Context(),
		frames: make(chan *s4wave_vm.WasiVmDisplayFrame, n),
		in:     make(chan *s4wave_vm.WasiVmDisplayInput, n),
	}
	scr := &testScreen{n: n, input: make(chan *s4wave_vm.WasiVmDisplayInput, 1)}
	done := make(chan error, 1)
	go func() { done <- serveScreen(t.Context(), strm.Send, strm.Recv, scr) }()

	// The first framesAhead frames arrive at once.
	next := func(want uint32) {
		t.Helper()
		if f := <-strm.frames; f.GetScreenWidth() != want {
			t.Fatalf("frame %d has width %d", want, f.GetScreenWidth())
		}
	}
	for i := range uint32(framesAhead) {
		next(i + 1)
	}

	// Input passes through while the next frame is held.
	key := &s4wave_vm.WasiVmDisplayInput{Body: &s4wave_vm.WasiVmDisplayInput_Key{Key: &s4wave_vm.WasiVmKeyEvent{Keysym: 'a', Down: true}}}
	strm.in <- key
	if got := <-scr.input; got != key {
		t.Fatalf("the screen got %v, want the key", got)
	}

	// Each later frame waits for the report of one drawn.
	drawn := &s4wave_vm.WasiVmDisplayInput{Body: &s4wave_vm.WasiVmDisplayInput_Drawn{Drawn: &s4wave_vm.WasiVmFrameDrawn{}}}
	for i := uint32(framesAhead); i < n; i++ {
		select {
		case f := <-strm.frames:
			t.Fatalf("frame %d arrived with %d undrawn", f.GetScreenWidth(), framesAhead)
		case <-time.After(20 * time.Millisecond):
		}
		strm.in <- drawn
		next(i + 1)
	}

	// The last frame returns once one more is drawn.
	strm.in <- drawn
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
