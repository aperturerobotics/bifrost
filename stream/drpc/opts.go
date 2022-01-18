package stream_drpc

import (
	"time"

	"github.com/aperturerobotics/bifrost/util/confparse"
	"storj.io/drpc/drpcconn"
	"storj.io/drpc/drpcmanager"
	"storj.io/drpc/drpcstream"
)

// Validate checks the ops.
func (o *DrpcOpts) Validate() error {
	if err := o.GetManagerOpts().Validate(); err != nil {
		return err
	}
	return nil
}

// BuildOpts converts to drpc conn opts.
func (o *DrpcOpts) BuildOpts() (drpcconn.Options, error) {
	opts := drpcconn.Options{}
	var err error
	opts.Manager, err = o.GetManagerOpts().BuildOpts()
	return opts, err
}

// Validate checks the ops.
func (o *ManagerOpts) Validate() error {
	if err := o.GetStreamOpts().Validate(); err != nil {
		return err
	}
	if _, err := o.ParseInactivityTimeout(); err != nil {
		return err
	}
	return nil
}

// BuildOpts converts to manager opts.
func (o *ManagerOpts) BuildOpts() (drpcmanager.Options, error) {
	streamOpts := o.GetStreamOpts().BuildOpts()
	opts := drpcmanager.Options{
		WriterBufferSize: int(o.GetWriterBufferSize()),
		Stream:           streamOpts,
	}
	var err error
	opts.InactivityTimeout, err = o.ParseInactivityTimeout()
	return opts, err
}

// ParseInactivityTimeout parses the inactivity timeout field.
func (o *ManagerOpts) ParseInactivityTimeout() (time.Duration, error) {
	return confparse.ParseDuration(o.GetInactivityTimeout())
}

// BuildOpts converts to stream opts.
func (o *StreamOpts) BuildOpts() drpcstream.Options {
	return drpcstream.Options{
		SplitSize: int(o.GetSplitSize()),
	}
}

// Validate checks the ops.
func (o *StreamOpts) Validate() error {
	// nothing to check
	return nil
}
