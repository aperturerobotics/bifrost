package logrw

import (
	"io"

	"github.com/sirupsen/logrus"
)

// LogWriteCloser wraps a io.WriterCloser with a logger.
type LogWriteCloser struct {
	*LogWriter
	underlying io.Closer
}

func NewLogWriteCloser(le *logrus.Entry, underlying io.WriteCloser) *LogWriteCloser {
	return &LogWriteCloser{
		LogWriter:  NewLogWriter(le, underlying),
		underlying: underlying,
	}
}

func (c *LogWriteCloser) Close() error {
	return c.underlying.Close()
}

// _ is a type assertion
var _ io.WriteCloser = ((*LogWriteCloser)(nil))
