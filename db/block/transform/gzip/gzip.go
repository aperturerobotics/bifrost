package transform_gzip

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"sync"

	kgzip "github.com/klauspost/compress/gzip"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	block_transform "github.com/s4wave/spacewave/db/block/transform"
)

// Gzip is the gzip compression step.
type Gzip struct {
	// level selects the encoder configuration.
	level int
	// writers shares reusable encoders with every transformer at this level.
	writers *sync.Pool
}

// gzipReader keeps the input reader stable so resetting it releases block data
// retained by both gzip and its underlying flate decompressor.
type gzipReader struct {
	// input releases the completed block when returned to the pool.
	input bytes.Reader
	// reader retains the decoder's reusable buffers.
	reader *kgzip.Reader
}

var (
	// gzipWriters shares codec storage across short-lived cursor transformers.
	// Each legal compression level has a separate pool.
	gzipWriters [gzip.BestCompression - gzip.HuffmanOnly + 1]sync.Pool
	// gzipReaders decodes independently of the encoder's compression level.
	gzipReaders = sync.Pool{New: func() any { return &gzipReader{} }}
	// decodeNative decodes a whole block in one call when a native decoder is
	// built in, and reports false to leave the block to the Go decoder.
	decodeNative func(data []byte) ([]byte, bool)
)

// NewGzip constructs the gzip compression step.
func NewGzip(c *Config) (*Gzip, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	level := c.EffectiveCompressionLevel()
	return &Gzip{level: level, writers: &gzipWriters[level-gzip.HuffmanOnly]}, nil
}

// EncodeBlock encodes the block according to the config.
// May reuse the same byte slice if possible.
func (g *Gzip) EncodeBlock(data []byte) ([]byte, error) {
	// Acquire a pooled gzip writer for this block output.
	var buf bytes.Buffer
	buf.Grow(len(data))
	wr, ok := g.writers.Get().(*gzip.Writer)
	if !ok {
		// The constructor validated the level before selecting this pool.
		wr, _ = gzip.NewWriterLevel(io.Discard, g.level)
	}
	wr.Reset(&buf)
	defer func() {
		// Keep compressor storage, but release the completed block's output buffer.
		wr.Reset(io.Discard)
		g.writers.Put(wr)
	}()

	// Compress the block and finish the gzip stream before exposing its bytes.
	if _, err := wr.Write(data); err != nil {
		_ = wr.Close()
		return nil, err
	}
	if err := wr.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecodeBlock decodes the block according to the config.
// May reuse the same byte slice if possible.
func (g *Gzip) DecodeBlock(data []byte) ([]byte, error) {
	// Decode through the native decoder when one is built in. The Go decoder
	// handles whatever it declines and reports the error for invalid input.
	if decodeNative != nil {
		if out, ok := decodeNative(data); ok {
			return out, nil
		}
	}

	// Acquire a pooled gzip reader and release block input after decoding.
	rd := gzipReaders.Get().(*gzipReader)
	rd.input.Reset(data)
	defer func() {
		rd.input.Reset(nil)
		if rd.reader != nil {
			_ = rd.reader.Close()
		}
		gzipReaders.Put(rd)
	}()

	// Initialize or reset the gzip decoder for the current block.
	var err error
	if rd.reader == nil {
		rd.reader, err = kgzip.NewReader(&rd.input)
	} else {
		err = rd.reader.Reset(&rd.input)
	}
	if err != nil {
		return nil, err
	}

	// Decode the block within the maximum permitted output size, sized once
	// from the trailer so the buffer never regrows. ReadFrom keeps MinRead
	// bytes free, which leaves room to observe the end of the stream.
	var buf bytes.Buffer
	buf.Grow(decodedSizeHint(data) + bytes.MinRead)
	lrd := &io.LimitedReader{R: rd.reader, N: block.MaxBlockSize + 1}
	if _, err := io.Copy(&buf, lrd); err != nil {
		return nil, err
	}

	// Reject gzip output that exceeds the block size limit.
	if lrd.N == 0 {
		return nil, errors.Errorf("gzip decoded block exceeds max block size: %d", block.MaxBlockSize)
	}

	return buf.Bytes(), nil
}

// maxDeflateRatio bounds how far deflate can expand its input.
const maxDeflateRatio = 1032

// decodedSizeHint returns the decoded size recorded in a gzip trailer, bounded
// by the block size limit and by what deflate could produce from data. The
// trailer is unverified until the stream ends, so it only sizes the output
// buffer.
func decodedSizeHint(data []byte) int {
	if len(data) < 18 {
		return 0
	}
	recorded := int(binary.LittleEndian.Uint32(data[len(data)-4:]))
	return min(recorded, len(data)*maxDeflateRatio, block.MaxBlockSize)
}

// _ is a type assertion
var _ block_transform.Step = (*Gzip)(nil)
