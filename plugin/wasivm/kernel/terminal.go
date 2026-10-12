package kernel

import "github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"

// initTermios sets the terminal's termios to the Linux cooked-mode defaults.
func (k *Kernel) initTermios() {
	// A console that is not a terminal has no termios.
	if !k.tty {
		return
	}

	// Translate input CR to NL and output NL to CRNL, run at 38400 baud
	// with 8-bit characters, and enable signals, canonical editing and echo.
	var tio termios
	tio.setFlags(icrnl, opost|onlcr, 0x04bf, 0x8a3b)
	copy(tio.cc(), defaultControlChars[:])
	k.setTermios(tio)
}

// defaultControlChars are the Linux default control characters.
var defaultControlChars = [...]byte{
	vintr:    0x03, // ^C
	vquit:    0x1c, // ^\
	verase:   0x7f, // DEL
	vkill:    0x15, // ^U
	veof:     0x04, // ^D
	vmin:     1,
	vstart:   0x11, // ^Q
	vstop:    0x13, // ^S
	vsusp:    0x1a, // ^Z
	vreprint: 0x12, // ^R
	vdiscard: 0x0f, // ^O
	vwerase:  0x17, // ^W
	vlnext:   0x16, // ^V
}

// termios returns the terminal's termios state.
func (k *Kernel) termios() termios {
	k.termMtx.Lock()
	defer k.termMtx.Unlock()
	return k.termIOS
}

// setTermios replaces the terminal's termios state.
func (k *Kernel) setTermios(tio termios) {
	k.termMtx.Lock()
	defer k.termMtx.Unlock()
	k.termIOS = tio
}

// syncHostTermMode puts a host terminal in raw mode while the termios
// ICANON bit is clear and restores it while the bit is set.
func (k *Kernel) syncHostTermMode() {
	if k.termFD < 0 {
		return
	}

	// Switch the host terminal when its mode differs.
	tio := k.termios()
	wantRaw := tio.lflag()&icanon == 0
	switch {
	case wantRaw && !k.termRaw.Load() && k.termMakeRaw != nil:
		state, err := k.termMakeRaw(k.termFD)
		if err == nil {
			k.termOldState = state
			k.termRaw.Store(true)
		}
	case !wantRaw && k.termRaw.Load():
		if k.termOldState != nil && k.termRestore != nil {
			_ = k.termRestore(k.termFD, k.termOldState)
		}
		k.termRaw.Store(false)
	}
}

// The wasi-libc termios struct holds the four mode words and the line
// discipline at the kernel's offsets, then 32 control characters and the
// input and output speeds.
const (
	// termiosCC is the offset of c_cc.
	termiosCC = 17
	// termiosNCCS is the length of c_cc.
	termiosNCCS = 32
	// termiosIspeed is the offset of c_ispeed.
	termiosIspeed = 52
	// termiosOspeed is the offset of c_ospeed.
	termiosOspeed = 56
	// termiosB38400 is the speed code for 38400 baud.
	termiosB38400 = 0o17
)

// writeWasiTermios writes the terminal's termios state to the wasi-libc
// termios struct at ptr. Control characters past the kernel's 19 are zero.
func (k *Kernel) writeWasiTermios(mem engine.Memory, ptr uint32) {
	// Write the modes, the control characters and the speeds.
	tio := k.termios()
	mem.Write(ptr, tio[:termiosCC])
	var cc [termiosNCCS]byte
	copy(cc[:], tio.cc())
	mem.Write(ptr+termiosCC, cc[:])
	mem.WriteUint32Le(ptr+termiosIspeed, termiosB38400)
	mem.WriteUint32Le(ptr+termiosOspeed, termiosB38400)
}

// readWasiTermios sets the terminal's termios state from the wasi-libc
// termios struct at ptr and syncs the host terminal mode to it.
func (k *Kernel) readWasiTermios(mem engine.Memory, ptr uint32) {
	// Read the modes and control characters, then apply them.
	buf, ok := mem.Read(ptr, uint32(len(termios{})))
	if !ok {
		return
	}
	var tio termios
	copy(tio[:], buf)
	k.setTermios(tio)
	k.syncHostTermMode()
}
