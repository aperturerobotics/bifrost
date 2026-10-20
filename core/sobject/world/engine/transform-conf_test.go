package sobject_world_engine

import (
	"context"
	"testing"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/s4wave/spacewave/db/block"
	block_transform "github.com/s4wave/spacewave/db/block/transform"
	transform_blockenc "github.com/s4wave/spacewave/db/block/transform/blockenc"
	transform_gzip "github.com/s4wave/spacewave/db/block/transform/gzip"
	transform_s2 "github.com/s4wave/spacewave/db/block/transform/s2"
	"github.com/s4wave/spacewave/db/util/blockenc"
	"github.com/s4wave/spacewave/db/world"
	world_block "github.com/s4wave/spacewave/db/world/block"
	world_block_tx "github.com/s4wave/spacewave/db/world/block/tx"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// TestSetTransformConfMovesWorldToS2 checks that a World written with gzip
// switches to s2, writes later blocks with s2, and still reads the blocks
// written before the switch.
func TestSetTransformConfMovesWorldToS2(t *testing.T) {
	// Write an object under the gzip transform.
	ctx := context.Background()
	pid := newProcessTestPeerID(t)
	c, so, headState := newProcessTestWorld(t, ctx)
	w := newReplayWorld(c, so)
	headState = processTransformTestOp(t, c, w, pid, 1, headState, transformTestCreateOp(t, headState, "before"))

	// Switch to s2.
	s2Conf := newStateTestTransformConfig(t, &transform_s2.Config{ReadGzip: true})
	headState = processTransformTestOp(t, c, w, pid, 2, headState, &SOWorldOp{Body: &SOWorldOp_SetTransformConf{
		SetTransformConf: &SetTransformConfOp{TransformConf: s2Conf},
	}})
	if !headState.GetHeadRef().GetTransformConf().EqualVT(s2Conf) {
		t.Fatal("head does not name the s2 transform")
	}

	// Write another object and store the head.
	headState = processTransformTestOp(t, c, w, pid, 3, headState, transformTestCreateOp(t, headState, "after"))
	if err := w.close(ctx); err != nil {
		t.Fatal(err.Error())
	}

	// The new root is s2: the gzip transform cannot read it.
	gzipXfrm, err := block_transform.NewTransformer(controller.ConstructOpts{}, c.sfs, newStateTestTransformConfig(t, &transform_gzip.Config{}))
	if err != nil {
		t.Fatal(err.Error())
	}
	_, bcs := block.NewTransaction(so.GetBlockStore(), gzipXfrm, headState.GetHeadRef().GetRootRef(), nil)
	if _, err := bcs.Unmarshal(ctx, world_block.NewWorldBlock); err == nil {
		t.Fatal("gzip transform read the root written after the switch")
	}

	// A fresh engine on the head reads both objects.
	blk, err := buildBlockEngine(ctx, c.le, c.bus, c.sfs, so, headState.GetHeadRef(), headState.GetHeadRef().GetTransformConf(), nil, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer blk.Release()
	read, err := blk.bengine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer read.Discard()
	for _, key := range []string{"before", "after"} {
		obj, found, err := read.GetObject(ctx, key)
		world.ReleaseObjectState(obj)
		if err != nil || !found {
			t.Fatalf("object %q found = %v, %v", key, found, err)
		}
	}
}

// TestValidateTransformSuccessor checks which transforms may follow the gzip
// World transform.
func TestValidateTransformSuccessor(t *testing.T) {
	// Name the successors replay accepts and rejects.
	prev := newStateTestTransformConfig(t, &transform_gzip.Config{})
	otherKey := &transform_blockenc.Config{BlockEnc: blockenc.DefaultBlockEnc, Key: make([]byte, 32)}
	otherKey.Key[0] = 1
	cases := []struct {
		name  string
		steps []config.Config
		ok    bool
	}{
		{"unchanged", []config.Config{&transform_gzip.Config{}}, true},
		{"s2 reading gzip", []config.Config{&transform_s2.Config{ReadGzip: true}}, true},
		{"s2 without read_gzip", []config.Config{&transform_s2.Config{}}, false},
		{"dropped step", nil, false},
		{"added step", []config.Config{&transform_gzip.Config{}, &transform_s2.Config{ReadGzip: true}}, false},
	}

	// Check each against the gzip transform.
	for _, tc := range cases {
		err := validateTransformSuccessor(prev, newStateTestTransformConfig(t, tc.steps...))
		if (err == nil) != tc.ok {
			t.Errorf("%s: err = %v; want ok %v", tc.name, err, tc.ok)
		}
	}

	// A changed key cannot read the old blocks, and neither can a transform
	// without encryption.
	changedKey, err := block_transform.NewConfig([]config.Config{&transform_s2.Config{ReadGzip: true}, otherKey})
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := validateTransformSuccessor(prev, changedKey); err == nil {
		t.Error("changed key accepted")
	}
	plain, err := block_transform.NewConfig([]config.Config{&transform_s2.Config{ReadGzip: true}})
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := validateTransformSuccessor(prev, plain); err == nil {
		t.Error("unencrypted transform accepted")
	}
}

// TestProcessSetTransformConfRejectsUninitializedWorld checks the operation
// needs a World.
func TestProcessSetTransformConfRejectsUninitializedWorld(t *testing.T) {
	op := &SetTransformConfOp{TransformConf: newStateTestTransformConfig(t, &transform_s2.Config{ReadGzip: true})}
	next, res, err := processSetTransformConfOp(logrus.NewEntry(logrus.New()), op, &InnerState{}, newProcessTestPeerID(t), 1)
	if err != nil || next != nil || res.GetSuccess() {
		t.Fatalf("process = %v, %v, %v; want rejection", next, res, err)
	}
}

// transformTestCreateOp returns an operation creating the object key.
func transformTestCreateOp(t *testing.T, headState *InnerState, key string) *SOWorldOp {
	t.Helper()
	tx, err := world_block_tx.NewTxCreateObject(key, headState.GetHeadRef().CloneVT())
	if err != nil {
		t.Fatal(err.Error())
	}
	return &SOWorldOp{Body: &SOWorldOp_ApplyTxOp{ApplyTxOp: &ApplyTxOp{Tx: tx}}}
}

// processTransformTestOp replays op on w and returns the accepted state.
func processTransformTestOp(t *testing.T, c *Controller, w *replayWorld, pid peer.ID, nonce uint64, headState *InnerState, op *SOWorldOp) *InnerState {
	// Encode and replay the operation.
	t.Helper()
	opData, err := op.MarshalVT()
	if err != nil {
		t.Fatal(err.Error())
	}
	next, res, err := c.processOp(context.Background(), c.le, w, opData, "test-op", pid, nonce, 0, headState)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !res.GetSuccess() {
		t.Fatalf("operation %d rejected: %s", nonce, res.GetErrorDetails().GetErrorMsg())
	}
	return next
}
