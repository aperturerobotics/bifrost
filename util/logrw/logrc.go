package logrw

import (
	"io"

	"github.com/sirupsen/logrus"
)

// LogReadCloser wraps a io.ReaderCloser with a logger.
type LogReadCloser struct {
	*LogReader
	underlying io.Closer
}

func NewLogReadCloser(le *logrus.Entry, underlying io.ReadCloser) *LogReadCloser {
	return &LogReadCloser{
		LogReader:  NewLogReader(le, underlying),
		underlying: underlying,
	}
}

func (c *LogReadCloser) Close() error {
	return c.underlying.Close()
}

// _ is a type assertion
var _ io.ReadCloser = ((*LogReadCloser)(nil))
