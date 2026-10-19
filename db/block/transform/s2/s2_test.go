package transform_s2

import (
	"bytes"
	"math/rand/v2"
	"testing"

	"github.com/klauspost/compress/s2"
	transform_gzip "github.com/s4wave/spacewave/db/block/transform/gzip"
)

func TestDecodeBlockReadsGzip(t *testing.T) {
	// Encode the same page as gzip, as stored before a switch, and as s2.
	data := bytes.Repeat([]byte("okra page entry "), 256)
	gz, err := transform_gzip.NewGzip(&transform_gzip.Config{})
	if err != nil {
		t.Fatal(err)
	}
	gzBlock, err := gz.EncodeBlock(data)
	if err != nil {
		t.Fatal(err)
	}

	// The step under test writes the s2 block.
	reader, err := NewS2(&Config{ReadGzip: true})
	if err != nil {
		t.Fatal(err)
	}
	s2Block, err := reader.EncodeBlock(data)
	if err != nil {
		t.Fatal(err)
	}

	// A step that reads gzip decodes both blocks.
	for _, block := range [][]byte{gzBlock, s2Block} {
		out, err := reader.DecodeBlock(block)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, data) {
			t.Fatal("decoded block differs from the input")
		}
	}

	// A step without the option rejects the gzip block.
	plain, err := NewS2(&Config{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plain.DecodeBlock(gzBlock); err == nil {
		t.Fatal("expected a plain s2 step to reject a gzip block")
	}
}

func TestS2RejectsGzipMagic(t *testing.T) {
	// No valid s2 block starts with the gzip magic, so routing those blocks to
	// gzip never misreads an s2 block.
	rng := rand.New(rand.NewPCG(1, 2))
	for range 10000 {
		block := []byte{0x1f, 0x8b, 0x08}
		suffix := make([]byte, rng.IntN(64))
		for i := range suffix {
			suffix[i] = byte(rng.Uint32())
		}
		block = append(block, suffix...)
		if _, err := s2.Decode(nil, block); err == nil {
			t.Fatalf("s2 decoded a block with the gzip magic: %x", block)
		}
	}
}
