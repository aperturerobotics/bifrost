package blob

import "errors"

var (
	// ErrRawBlobSizeMismatch is returned if the raw blob size field does not match the data len.
	ErrRawBlobSizeMismatch = errors.New("raw blob size must match data len")
	// ErrEmptyChunk is returned if a empty chunk was found (invalid).
	ErrEmptyChunk = errors.New("empty chunk is invalid")
	// ErrOutOfSequenceChunk is returned if a chunk was out-of-sequence (invalid size or start).
	ErrOutOfSequenceChunk = errors.New("invalid chunk sequence")
	// ErrUnknownBlobType is returned for a blob type that is not recognized.
	ErrUnknownBlobType = errors.New("unknown blob type")
	// ErrUnknownChunkerType is returned for a chunker type that is not recognized.
	ErrUnknownChunkerType = errors.New("unknown chunker type")
	// ErrNoChunkCursor is returned when a chunk has no cursor to hold its data
	// block, such as a blob built at a nil cursor.
	ErrNoChunkCursor = errors.New("chunk data needs a cursor: build the blob at a non-nil cursor")
)
