package kernel

import "encoding/binary"

// termios is a Linux termios struct in the asm-generic layout: the input,
// output, control and local mode words, the line discipline byte and 19
// control characters.
type termios [36]byte

// iflag returns the input modes.
func (t *termios) iflag() uint32 {
	return binary.LittleEndian.Uint32(t[0:4])
}

// oflag returns the output modes.
func (t *termios) oflag() uint32 {
	return binary.LittleEndian.Uint32(t[4:8])
}

// lflag returns the local modes.
func (t *termios) lflag() uint32 {
	return binary.LittleEndian.Uint32(t[12:16])
}

// The indexes of the control characters.
const (
	vintr    = 0
	vquit    = 1
	verase   = 2
	vkill    = 3
	veof     = 4
	vmin     = 6
	vstart   = 8
	vstop    = 9
	vsusp    = 10
	vreprint = 12
	vdiscard = 13
	vwerase  = 14
	vlnext   = 15
)

// cc returns the control characters, indexed by vintr and the other v
// constants.
func (t *termios) cc() []byte {
	return t[17:36]
}

// setFlags sets the input, output, control and local modes.
func (t *termios) setFlags(iflag, oflag, cflag, lflag uint32) {
	binary.LittleEndian.PutUint32(t[0:4], iflag)
	binary.LittleEndian.PutUint32(t[4:8], oflag)
	binary.LittleEndian.PutUint32(t[8:12], cflag)
	binary.LittleEndian.PutUint32(t[12:16], lflag)
}
