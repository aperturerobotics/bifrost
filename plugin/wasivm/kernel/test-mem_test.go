package kernel

import (
	"context"
	"encoding/binary"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// wasmPageSize is the size of a WebAssembly memory page.
const wasmPageSize = 65536

// memMemory is an engine.Memory over a byte slice, so a test can call a
// syscall that reads guest memory without running a module.
type memMemory struct {
	buf []byte
}

func newMemMemory(size uint32) *memMemory {
	return &memMemory{buf: make([]byte, size)}
}

func (m *memMemory) inBounds(offset, length uint32) bool {
	return uint64(offset)+uint64(length) <= uint64(len(m.buf))
}

func (m *memMemory) Read(offset, length uint32) ([]byte, bool) {
	if !m.inBounds(offset, length) {
		return nil, false
	}
	out := make([]byte, length)
	copy(out, m.buf[offset:offset+length])
	return out, true
}

func (m *memMemory) ReadByteAt(offset uint32) (byte, bool) {
	if !m.inBounds(offset, 1) {
		return 0, false
	}
	return m.buf[offset], true
}

func (m *memMemory) ReadUint32Le(offset uint32) (uint32, bool) {
	if !m.inBounds(offset, 4) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(m.buf[offset:]), true
}

func (m *memMemory) ReadUint64Le(offset uint32) (uint64, bool) {
	if !m.inBounds(offset, 8) {
		return 0, false
	}
	return binary.LittleEndian.Uint64(m.buf[offset:]), true
}

func (m *memMemory) Write(offset uint32, data []byte) bool {
	if !m.inBounds(offset, uint32(len(data))) {
		return false
	}
	copy(m.buf[offset:], data)
	return true
}

func (m *memMemory) WriteByteAt(offset uint32, val byte) bool {
	if !m.inBounds(offset, 1) {
		return false
	}
	m.buf[offset] = val
	return true
}

func (m *memMemory) WriteUint32Le(offset, val uint32) bool {
	if !m.inBounds(offset, 4) {
		return false
	}
	binary.LittleEndian.PutUint32(m.buf[offset:], val)
	return true
}

func (m *memMemory) WriteUint64Le(offset uint32, val uint64) bool {
	if !m.inBounds(offset, 8) {
		return false
	}
	binary.LittleEndian.PutUint64(m.buf[offset:], val)
	return true
}

// writeCString writes s at ptr as a NUL-terminated string.
func (m *memMemory) writeCString(ptr uint32, s string) {
	m.Write(ptr, append([]byte(s), 0))
}

func (m *memMemory) Size() uint32 { return uint32(len(m.buf)) }

func (m *memMemory) Grow(pages uint32) (uint32, bool) {
	prev := uint32(len(m.buf)) / wasmPageSize
	m.buf = append(m.buf, make([]byte, pages*wasmPageSize)...)
	return prev, true
}

// memInstance is an engine.Instance whose memory is a memMemory.
type memInstance struct {
	mem *memMemory
}

func (i memInstance) Memory() engine.Memory                     { return i.mem }
func (i memInstance) ExportedFunction(_ string) engine.Function { return nil }
func (i memInstance) ExportedGlobal(_ string) engine.Global     { return nil }
func (i memInstance) Close(_ context.Context) error             { return nil }

// newMemProcess returns PID 1 of a minimal kernel, with memory of size bytes.
func newMemProcess(size uint32) (*Process, *memMemory) {
	// Register the process, then give it memory.
	k := &Kernel{procs: newProcessTable()}
	p := newProcess(k, 1, nil)
	mem := newMemMemory(size)
	p.inst = memInstance{mem: mem}
	k.procs.set(1, p)
	return p, mem
}
