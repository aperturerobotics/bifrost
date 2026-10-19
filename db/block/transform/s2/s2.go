package transform_s2

import (
	"github.com/klauspost/compress/s2"
	block_transform "github.com/s4wave/spacewave/db/block/transform"
	transform_gzip "github.com/s4wave/spacewave/db/block/transform/gzip"
)

// S2 is the S2 compression step.
type S2 struct {
	better bool
	best   bool
	// gzip decodes blocks with the gzip magic, when the config reads gzip.
	gzip *transform_gzip.Gzip
}

// NewS2 constructs the s2 compress step.
func NewS2(c *Config) (*S2, error) {
	s := &S2{
		better: c.GetBetter(),
		best:   c.GetBest(),
	}

	// Build the gzip decoder for blocks written before a switch to s2.
	if c.GetReadGzip() {
		g, err := transform_gzip.NewGzip(&transform_gzip.Config{})
		if err != nil {
			return nil, err
		}
		s.gzip = g
	}
	return s, nil
}

// EncodeBlock encodes the block according to the config.
// May reuse the same byte slice if possible.
func (s *S2) EncodeBlock(data []byte) ([]byte, error) {
	switch {
	case s.best:
		return s2.EncodeBest(nil, data), nil
	case s.better:
		return s2.EncodeBetter(nil, data), nil
	default:
		return s2.Encode(nil, data), nil
	}
}

// DecodeBlock decodes the block according to the config.
// May reuse the same byte slice if possible.
func (s *S2) DecodeBlock(data []byte) ([]byte, error) {
	// s2 output never starts with the gzip magic: 1f is a 31 byte length,
	// and 8b is then a copy before any literal, which s2 rejects.
	if s.gzip != nil && isGzip(data) {
		return s.gzip.DecodeBlock(data)
	}
	return s2.Decode(nil, data)
}

// isGzip reports whether data starts with the gzip magic and deflate method.
func isGzip(data []byte) bool {
	return len(data) >= 3 && data[0] == 0x1f && data[1] == 0x8b && data[2] == 8
}

// _ is a type assertion
var _ block_transform.Step = (*S2)(nil)
