package blockcompress

import (
	"io"

	"github.com/pkg/errors"
)

// BuildCompressStream builds the compression stream for the given block crypto type.
func (b BlockCompress) BuildCompressStream(w io.ReadWriteCloser) (io.ReadWriteCloser, error) {
	switch b {
	case BlockCompress_BlockCompress_NONE:
		return w, nil
	case BlockCompress_BlockCompress_SNAPPY:
		return NewSnappyStream(w), nil
	case BlockCompress_BlockCompress_S2:
		return NewS2Stream(w), nil
	case BlockCompress_BlockCompress_LZ4:
		return NewLz4Stream(w), nil
	case BlockCompress_BlockCompress_ZSTD:
		return NewZStdStream(w), nil
	default:
		return nil, errors.Errorf("invalid block compress type: %s", b.String())
	}
}
