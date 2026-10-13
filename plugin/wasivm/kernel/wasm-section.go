package kernel

// wasmPageSize is the size of a WebAssembly memory page.
const wasmPageSize = 65536

// Section ids of the WebAssembly binary format.
const (
	// wasmSectionCustom holds named data such as the name section.
	wasmSectionCustom byte = 0
	// wasmSectionType holds the function types.
	wasmSectionType byte = 1
	// wasmSectionImport holds the imports.
	wasmSectionImport byte = 2
	// wasmSectionFunction holds the type of each defined function.
	wasmSectionFunction byte = 3
	// wasmSectionTable holds the defined tables.
	wasmSectionTable byte = 4
	// wasmSectionGlobal holds the defined globals.
	wasmSectionGlobal byte = 6
	// wasmSectionExport holds the exports.
	wasmSectionExport byte = 7
	// wasmSectionCode holds the body of each defined function.
	wasmSectionCode byte = 10
)

// wasmSectionOrder maps each known section id to its required position. The
// tag and data count sections are numbered out of order.
var wasmSectionOrder = map[byte]int{
	1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 13: 6, 6: 7, 7: 8, 8: 9, 9: 10, 12: 11, 10: 12, 11: 13,
}

// wasmSection is one section of a WebAssembly binary.
type wasmSection struct {
	// id identifies the section kind.
	id byte
	// payload is the section contents after its size.
	payload []byte
}

// after reports whether s must follow a section with id in a binary. Custom
// sections may appear anywhere, so they follow nothing.
func (s wasmSection) after(id byte) bool {
	if s.id == wasmSectionCustom {
		return false
	}
	return wasmSectionOrder[s.id] > wasmSectionOrder[id]
}
