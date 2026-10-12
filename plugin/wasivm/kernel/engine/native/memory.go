package native

import (
	"github.com/tetratelabs/wazero/api"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// Memory is the linear memory of a wazero guest.
type Memory struct {
	// mem is the wazero memory.
	mem api.Memory
}

// Read returns length bytes at offset. The slice aliases guest memory.
func (m *Memory) Read(offset, length uint32) ([]byte, bool) {
	return m.mem.Read(offset, length)
}

// ReadByteAt returns the byte at offset.
func (m *Memory) ReadByteAt(offset uint32) (byte, bool) {
	return m.mem.ReadByte(offset)
}

// ReadUint32Le returns the little-endian uint32 at offset.
func (m *Memory) ReadUint32Le(offset uint32) (uint32, bool) {
	return m.mem.ReadUint32Le(offset)
}

// ReadUint64Le returns the little-endian uint64 at offset.
func (m *Memory) ReadUint64Le(offset uint32) (uint64, bool) {
	return m.mem.ReadUint64Le(offset)
}

// Write copies data to offset.
func (m *Memory) Write(offset uint32, data []byte) bool {
	return m.mem.Write(offset, data)
}

// WriteByteAt writes one byte at offset.
func (m *Memory) WriteByteAt(offset uint32, val byte) bool {
	return m.mem.WriteByte(offset, val)
}

// WriteUint32Le writes a little-endian uint32 at offset.
func (m *Memory) WriteUint32Le(offset uint32, val uint32) bool {
	return m.mem.WriteUint32Le(offset, val)
}

// WriteUint64Le writes a little-endian uint64 at offset.
func (m *Memory) WriteUint64Le(offset uint32, val uint64) bool {
	return m.mem.WriteUint64Le(offset, val)
}

// Size returns the memory size in bytes.
func (m *Memory) Size() uint32 {
	return m.mem.Size()
}

// Grow adds pages of 64 KiB and returns the previous size in pages.
func (m *Memory) Grow(pages uint32) (uint32, bool) {
	return m.mem.Grow(pages)
}

// _ is a type assertion
var _ engine.Memory = (*Memory)(nil)
