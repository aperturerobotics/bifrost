package kernel

import "github.com/pkg/errors"

// errWasmMalformed is returned for a guest binary the kernel cannot parse.
var errWasmMalformed = errors.New("malformed wasm binary")

// wasmReader decodes WebAssembly binary encoding front to back. The first
// failure sticks: later reads return zero values, and err reports it.
type wasmReader struct {
	// dat is the encoded input.
	dat []byte
	// off is the offset of the next unread byte.
	off int
	// err is the first decoding failure.
	err error
}

// done reports whether the input is exhausted or a read failed.
func (r *wasmReader) done() bool {
	return r.err != nil || r.off >= len(r.dat)
}

// rest returns the unread input.
func (r *wasmReader) rest() []byte {
	if r.err != nil {
		return nil
	}
	return r.dat[r.off:]
}

// readByte reads one byte.
func (r *wasmReader) readByte() byte {
	if r.err != nil || r.off >= len(r.dat) {
		r.fail()
		return 0
	}
	b := r.dat[r.off]
	r.off++
	return b
}

// readU32 reads an unsigned LEB128 value. It also steps over a signed one.
func (r *wasmReader) readU32() uint32 {
	var val uint32
	for shift := 0; shift < 35; shift += 7 {
		b := r.readByte()
		val |= uint32(b&0x7f) << shift
		if b&0x80 == 0 {
			return val
		}
	}
	r.fail()
	return 0
}

// readBytes reads n bytes. The slice aliases the input.
func (r *wasmReader) readBytes(n uint32) []byte {
	// Check the bytes are there, then take them.
	end := r.off + int(n)
	if r.err != nil || end < r.off || end > len(r.dat) {
		r.fail()
		return nil
	}
	dat := r.dat[r.off:end]
	r.off = end
	return dat
}

// readName reads a length-prefixed name.
func (r *wasmReader) readName() string {
	return string(r.readBytes(r.readU32()))
}

// skipLEB steps over a LEB128 value of any width.
func (r *wasmReader) skipLEB() {
	for !r.done() {
		if r.readByte()&0x80 == 0 {
			return
		}
	}
}

// skipConstExpr steps over a constant expression and its end.
func (r *wasmReader) skipConstExpr() {
	for !r.done() {
		switch r.readByte() {
		case 0x0b:
			return
		case 0x41, 0x42, 0x23, 0xd2:
			// i32.const, i64.const, global.get and ref.func.
			r.skipLEB()
		case 0x43:
			r.readBytes(4)
		case 0x44:
			r.readBytes(8)
		case 0xd0:
			// ref.null.
			r.readByte()
		case 0x6a, 0x6b, 0x6c, 0x7c, 0x7d, 0x7e:
			// The extended constant arithmetic takes no immediates.
		default:
			r.fail()
		}
	}
}

// skipLimits steps over the limits of a table or memory type.
func (r *wasmReader) skipLimits() {
	flags := r.readByte()
	r.readU32()
	if flags&0x01 != 0 {
		r.readU32()
	}
}

// fail records a malformed input.
func (r *wasmReader) fail() {
	if r.err == nil {
		r.err = errWasmMalformed
	}
}
