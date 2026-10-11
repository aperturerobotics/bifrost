package s4wave_world

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	resource_client "github.com/s4wave/spacewave/bldr/resource/client"
	"github.com/s4wave/spacewave/db/coord"
)

// Tx represents a transaction against the world state.
// Tx implements the world state transaction interfaces. It provides the full
// WorldState read and write interface it inherits, and the Commit and Discard
// operations of tx.Tx.
//
// Tx maintains state across multiple RPC calls, enabling complex multi-step
// operations within a single transaction. Always call Discard() when done.
//
// Concurrent calls to WorldState functions should be supported.
type Tx struct {
	*WorldState
	txService SRPCTxResourceServiceClient
}

// NewTx creates a new Tx resource wrapper.
func NewTx(client *resource_client.Client, ref resource_client.ResourceRef, readOnly bool) (*Tx, error) {
	// Wrap the transaction resource with its World state interface.
	ws, err := NewWorldState(client, ref, readOnly)
	if err != nil {
		return nil, err
	}

	// Obtain the transaction client for commit and discard operations.
	srpcClient, err := ref.GetClient()
	if err != nil {
		return nil, err
	}

	return &Tx{
		WorldState: ws,
		txService:  NewSRPCTxResourceServiceClient(srpcClient),
	}, nil
}

// Commit commits the transaction.
// After commit, the transaction should be discarded and its resource released.
// Commit waits for the server's durable publication, not merely admission. A
// concurrent Engine.NewTransaction may prepare the next revision after this
// transaction seals but before this call completes. Keep every outstanding
// Commit result and join them before reporting success.
//
// A ctx that is already done returns its error and writes nothing. The request
// ignores ctx's cancellation from the call on, because canceling an RPC does
// not roll back an admitted write: a caller's deadline never abandons a commit
// that started, and a nil result means the write is placed. Closing the
// transport still ends the call, and its outcome is then unknown.
//
// A commit rejected because another writer advanced the World returns an error
// matching coord.ErrStaleGeneration; the caller may retry from a new
// transaction.
func (tx *Tx) Commit(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := tx.txService.Commit(context.WithoutCancel(ctx), &CommitRequest{})
	return CommitError(err)
}

// CommitError restores the stale-generation classification that the RPC
// boundary flattens to text in a Commit error.
func CommitError(err error) error {
	// Preserve a successful transaction commit.
	if err == nil {
		return nil
	}

	// Recognize a stale-generation error flattened by the RPC boundary.
	msg, stale := strings.CutSuffix(err.Error(), coord.ErrStaleGeneration.Error())
	if !stale {
		return err
	}

	// Restore the stale-generation classification with its remote message.
	msg = strings.TrimSuffix(msg, ": ")
	if msg == "" {
		return coord.ErrStaleGeneration
	}
	return errors.Wrap(coord.ErrStaleGeneration, msg)
}

// Discard discards the transaction without committing changes.
// Always call this when done with the transaction.
func (tx *Tx) Discard(ctx context.Context) error {
	_, err := tx.txService.Discard(ctx, &DiscardRequest{})
	return err
}
