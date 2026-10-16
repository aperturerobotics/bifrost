package git_world

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// NewGitCloneOp constructs a new GitCloneOp.
func NewGitCloneOp(op *GitCloneOp) *GitCloneOp {
	return op
}

// Validate checks the clone operation.
func (o *GitCloneOp) Validate() error {
	if o.GetObjectKey() == "" {
		return world.ErrEmptyObjectKey
	}
	if err := o.GetCloneOpts().Validate(); err != nil {
		return errors.Wrap(err, "clone_opts")
	}
	if !o.GetDisableCheckout() {
		if wt := o.GetCreateWorktree(); wt != nil {
			if err := wt.Validate(); err != nil {
				return errors.Wrap(err, "create_worktree")
			}
		}
	}
	return nil
}

// ApplyWorldOp applies the operation as a world operation.
func (o *GitCloneOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	worldHandle world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	cloneOpts := o.GetCloneOpts()
	if o.GetDisableCheckout() {
		cloneOpts = cloneOpts.CloneVT()
		cloneOpts.DisableCheckout = true
	}
	_, err = GitClone(
		ctx,
		worldHandle,
		o.GetObjectKey(),
		sender,
		cloneOpts,
		nil,
		nil,
		o.GetCreateWorktree(),
		o.GetTimestamp(),
	)
	return false, err
}

// _ is a type assertion
var _ world.Operation = (*GitCloneOp)(nil)
