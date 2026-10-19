//go:build libdeflate && cgo

// Package libdeflate decodes gzip blocks with libdeflate, whose sources are
// vendored here by vendor-libdeflate.py.
//
// Cross builds with zig cc need CGO_CFLAGS=-mevex512: zig disables evex512 on
// baseline x86-64, and libdeflate's AVX-512 CRC kernels need it to compile.
// Runtime CPU detection still selects the kernel the host supports.
package libdeflate

/*
#cgo CFLAGS: -O2
#include "libdeflate.h"
*/
import "C"

import (
	"runtime"
	"sync"
	"unsafe"
)

// decompressor owns one libdeflate decompressor, freed when collected.
type decompressor struct {
	// d is the native decompressor state.
	d *C.struct_libdeflate_decompressor
}

// decompressors shares native decompressors between goroutines.
var decompressors = sync.Pool{New: func() any {
	d := &decompressor{d: C.libdeflate_alloc_decompressor()}
	runtime.AddCleanup(d, func(p *C.struct_libdeflate_decompressor) {
		C.libdeflate_free_decompressor(p)
	}, d.d)
	return d
}}

// DecodeGzip decodes data, which must be one gzip member, into exactly size
// bytes. It reports false when data is anything else or fails to decode.
func DecodeGzip(data []byte, size int) ([]byte, bool) {
	// Allocate the exact output, with a valid pointer for an empty one.
	if len(data) == 0 || size < 0 {
		return nil, false
	}
	out := make([]byte, size, max(size, 1))

	// Decode in one call with a pooled decompressor.
	d := decompressors.Get().(*decompressor)
	var in, n C.size_t
	res := C.libdeflate_gzip_decompress_ex(
		d.d,
		unsafe.Pointer(unsafe.SliceData(data)), C.size_t(len(data)),
		unsafe.Pointer(unsafe.SliceData(out[:cap(out)])), C.size_t(size),
		&in, &n,
	)
	decompressors.Put(d)

	// Accept only a single member that fills the expected output.
	if res != C.LIBDEFLATE_SUCCESS || int(in) != len(data) || int(n) != size {
		return nil, false
	}
	return out, true
}
