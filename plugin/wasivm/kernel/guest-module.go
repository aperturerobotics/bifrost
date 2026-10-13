package kernel

import (
	"bytes"
	"slices"

	"github.com/pkg/errors"
)

// Export kinds of the WebAssembly binary format.
const (
	// wasmExportFunc exports a function.
	wasmExportFunc byte = 0x00
	// wasmExportGlobal exports a global.
	wasmExportGlobal byte = 0x03
)

// wasmMagic is the header of every WebAssembly binary: the magic number and
// format version 1.
var wasmMagic = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

// guestModule is a guest binary split into sections, so the kernel can add
// the exports it relies on before compiling it. Edits only append types,
// functions and exports, so every existing index stays valid.
type guestModule struct {
	// sections are the module's sections in binary order.
	sections []wasmSection
}

// parseGuestModule splits a WebAssembly binary into its sections.
func parseGuestModule(wasm []byte) (*guestModule, error) {
	// Check the header.
	if !bytes.HasPrefix(wasm, wasmMagic) {
		return nil, errors.Wrap(errWasmMalformed, "missing wasm header")
	}

	// Read each section's id and payload.
	r := &wasmReader{dat: wasm, off: len(wasmMagic)}
	m := &guestModule{}
	for !r.done() {
		id := r.readByte()
		payload := r.readBytes(r.readU32())
		m.sections = append(m.sections, wasmSection{id: id, payload: payload})
	}
	return m, r.err
}

// encode returns the module as a WebAssembly binary.
func (m *guestModule) encode() []byte {
	out := bytes.Clone(wasmMagic)
	for _, s := range m.sections {
		out = append(out, s.id)
		out = appendWasmU32(out, uint32(len(s.payload)))
		out = append(out, s.payload...)
	}
	return out
}

// payload returns the contents of the section with id, or nil when the
// module has none.
func (m *guestModule) payload(id byte) []byte {
	for _, s := range m.sections {
		if s.id == id {
			return s.payload
		}
	}
	return nil
}

// setPayload replaces the contents of the section with id, adding the
// section in binary order when the module has none.
func (m *guestModule) setPayload(id byte, payload []byte) {
	// Replace an existing section.
	for i := range m.sections {
		if m.sections[i].id == id {
			m.sections[i].payload = payload
			return
		}
	}

	// Insert a new one before the first section that follows it.
	at := len(m.sections)
	for i, s := range m.sections {
		if s.after(id) {
			at = i
			break
		}
	}
	m.sections = slices.Insert(m.sections, at, wasmSection{id: id, payload: payload})
}

// wasmExport is the kind and index of an exported item.
type wasmExport struct {
	kind  byte
	index uint32
}

// exports returns the module's exports by name.
func (m *guestModule) exports() (map[string]wasmExport, error) {
	// An absent section exports nothing.
	r := &wasmReader{dat: m.payload(wasmSectionExport)}
	exports := make(map[string]wasmExport)
	if r.done() {
		return exports, nil
	}

	// Read each export's name, kind and index.
	for range r.readU32() {
		name := r.readName()
		exports[name] = wasmExport{kind: r.readByte(), index: r.readU32()}
	}
	return exports, r.err
}

// addExport exports the item of kind at index under name.
func (m *guestModule) addExport(name string, kind byte, index uint32) error {
	// Append the entry to the export section.
	entry := appendWasmName(nil, name)
	entry = append(entry, kind)
	entry = appendWasmU32(entry, index)
	payload, _, err := appendWasmVector(m.payload(wasmSectionExport), entry)
	if err != nil {
		return err
	}
	m.setPayload(wasmSectionExport, payload)
	return nil
}

// funcType returns the index of the function type with params and results,
// adding the type when the module lacks it.
func (m *guestModule) funcType(params, results []byte) (uint32, error) {
	// Encode the type.
	entry := []byte{0x60}
	entry = appendWasmU32(entry, uint32(len(params)))
	entry = append(entry, params...)
	entry = appendWasmU32(entry, uint32(len(results)))
	entry = append(entry, results...)

	// Reuse a matching type. Value types of the supported features are one
	// byte each.
	vec := m.payload(wasmSectionType)
	r := &wasmReader{dat: vec}
	if !r.done() {
		for i := range r.readU32() {
			start := r.off
			if r.readByte() != 0x60 {
				return 0, errors.Wrap(errWasmMalformed, "unsupported type form")
			}
			r.readBytes(r.readU32())
			r.readBytes(r.readU32())
			if r.err == nil && bytes.Equal(vec[start:r.off], entry) {
				return i, nil
			}
		}
		if r.err != nil {
			return 0, r.err
		}
	}

	// Append a new type.
	payload, index, err := appendWasmVector(vec, entry)
	if err != nil {
		return 0, err
	}
	m.setPayload(wasmSectionType, payload)
	return index, nil
}

// addFunction defines a function of type typeIndex with body, the encoded
// locals and instructions, and returns its function index.
func (m *guestModule) addFunction(typeIndex uint32, body []byte) (uint32, error) {
	// The new function follows the imported and defined functions.
	imported, err := m.imports()
	if err != nil {
		return 0, err
	}
	funcs, defined, err := appendWasmVector(m.payload(wasmSectionFunction), appendWasmU32(nil, typeIndex))
	if err != nil {
		return 0, err
	}

	// Record its type and body.
	code := appendWasmU32(nil, uint32(len(body)))
	code, _, err = appendWasmVector(m.payload(wasmSectionCode), append(code, body...))
	if err != nil {
		return 0, err
	}
	m.setPayload(wasmSectionFunction, funcs)
	m.setPayload(wasmSectionCode, code)
	return imported.funcs + defined, nil
}

// wasmImports counts the items a module imports by kind. Imported items
// take the first indices of their kind.
type wasmImports struct {
	funcs   uint32
	tables  uint32
	globals uint32
}

// imports counts the functions, tables and globals the module imports.
func (m *guestModule) imports() (wasmImports, error) {
	// Read the import section, which a module without imports lacks.
	r := &wasmReader{dat: m.payload(wasmSectionImport)}
	var n wasmImports
	if r.done() {
		return n, nil
	}

	// Count each import by its kind.
	for range r.readU32() {
		r.readName()
		r.readName()
		switch r.readByte() {
		case 0x00:
			n.funcs++
			r.readU32()
		case 0x01:
			n.tables++
			r.readByte()
			r.skipLimits()
		case 0x02:
			r.skipLimits()
		case 0x03:
			n.globals++
			r.readBytes(2)
		case 0x04:
			r.readByte()
			r.readU32()
		default:
			r.fail()
		}
	}
	return n, r.err
}

// mutableGlobals returns the index of each mutable global the module
// defines.
func (m *guestModule) mutableGlobals() ([]uint32, error) {
	// Defined globals follow the imported ones.
	imported, err := m.imports()
	if err != nil {
		return nil, err
	}
	r := &wasmReader{dat: m.payload(wasmSectionGlobal)}
	if r.done() {
		return nil, nil
	}

	// Read each global's type, mutability and initializer.
	var out []uint32
	for i := range r.readU32() {
		r.readByte()
		if r.readByte() == 0x01 {
			out = append(out, imported.globals+i)
		}
		r.skipConstExpr()
	}
	return out, r.err
}

// hasTable reports whether the module imports or defines a table.
func (m *guestModule) hasTable() (bool, error) {
	imported, err := m.imports()
	if err != nil || imported.tables != 0 {
		return imported.tables != 0, err
	}
	r := &wasmReader{dat: m.payload(wasmSectionTable)}
	return !r.done() && r.readU32() != 0, r.err
}

// namedGlobal returns the index of the global the name section calls name.
func (m *guestModule) namedGlobal(name string) (uint32, bool) {
	// Find the custom section called "name".
	var names *wasmReader
	for _, s := range m.sections {
		r := &wasmReader{dat: s.payload}
		if s.id == wasmSectionCustom && r.readName() == "name" {
			names = r
			break
		}
	}
	if names == nil {
		return 0, false
	}

	// Find the global names subsection and the entry for name.
	for !names.done() {
		id := names.readByte()
		sub := &wasmReader{dat: names.readBytes(names.readU32())}
		if id != 7 {
			continue
		}
		for range sub.readU32() {
			index := sub.readU32()
			if sub.readName() == name && sub.err == nil {
				return index, true
			}
		}
	}
	return 0, false
}

// appendWasmVector returns the vector vec with entry appended, and the index
// of entry. An empty vec is an empty vector.
func appendWasmVector(vec, entry []byte) ([]byte, uint32, error) {
	// Read the count.
	r := &wasmReader{dat: vec}
	var count uint32
	if !r.done() {
		count = r.readU32()
	}
	if r.err != nil {
		return nil, 0, r.err
	}

	// Write the new count, the old entries and entry.
	out := appendWasmU32(nil, count+1)
	out = append(out, r.rest()...)
	return append(out, entry...), count, nil
}

// appendWasmName appends a length-prefixed name.
func appendWasmName(dst []byte, name string) []byte {
	dst = appendWasmU32(dst, uint32(len(name)))
	return append(dst, name...)
}

// appendWasmU32 appends val as unsigned LEB128.
func appendWasmU32(dst []byte, val uint32) []byte {
	for val >= 0x80 {
		dst = append(dst, byte(val)|0x80)
		val >>= 7
	}
	return append(dst, byte(val))
}
