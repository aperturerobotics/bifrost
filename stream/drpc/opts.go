package stream_drpc

import (
	"storj.io/drpc/drpcconn"
	"storj.io/drpc/drpcmanager"
	"storj.io/drpc/drpcstream"
)

// BuildOpts converts to drpc conn opts.
func (o *DrpcOpts) BuildOpts() drpcconn.Options {
	return drpcconn.Options{
		Manager: o.GetManagerOpts().BuildOpts(),
	}
}

// BuildOpts converts to manager opts.
func (o *ManagerOpts) BuildOpts() drpcmanager.Options {
	streamOpts := o.GetStreamOpts().BuildOpts()
	return drpcmanager.Options{
		WriterBufferSize: int(o.GetWriterBufferSize()),
		Stream:           streamOpts,
	}
}

// BuildOpts converts to stream opts.
func (o *StreamOpts) BuildOpts() drpcstream.Options {
	return drpcstream.Options{
		SplitSize: int(o.GetSplitSize()),
	}
}
