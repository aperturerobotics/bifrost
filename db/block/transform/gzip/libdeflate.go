//go:build libdeflate && cgo

package transform_gzip

import (
	"encoding/binary"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/block/transform/gzip/libdeflate"
)

// init routes block decoding through libdeflate.
func init() {
	decodeNative = decodeLibdeflate
}

// decodeLibdeflate decodes a block sized by its trailer in one libdeflate call.
// It declines a trailer beyond the block size limit or what deflate could
// produce, so the Go decoder reports those blocks.
func decodeLibdeflate(data []byte) ([]byte, bool) {
	if len(data) < 18 {
		return nil, false
	}
	size := int(binary.LittleEndian.Uint32(data[len(data)-4:]))
	if size > block.MaxBlockSize || size > len(data)*maxDeflateRatio {
		return nil, false
	}
	return libdeflate.DecodeGzip(data, size)
}
