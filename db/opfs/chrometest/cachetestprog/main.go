//go:build js

package main

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"syscall/js"
	"time"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/opfs"
	"github.com/s4wave/spacewave/db/volume/js/opfs/engine"
)

// config selects one cache probe worker and its shared workload dimensions.
type config struct {
	// scenario selects the worker action.
	scenario string
	// root names the shared OPFS volume directory.
	root string
	// worker identifies this worker among the concurrent publishers.
	worker int
	// workers is the number of concurrent publishers expected by a reader.
	workers int
	// iterations is the number of batches published by each writer.
	iterations int
	// batch is the number of blocks in each published generation.
	batch int
}

// blockEvent describes one published batch or writer completion.
type blockEvent struct {
	// typ identifies the event kind sent through BroadcastChannel.
	typ string
	// worker identifies the publisher that emitted the event.
	worker int
	// iteration identifies a published batch within one writer.
	iteration int
}

// blockEventSub owns one BroadcastChannel subscription and its Go event queue.
type blockEventSub struct {
	// ch buffers events so the JavaScript callback never blocks.
	ch chan blockEvent
	// bc is the underlying browser BroadcastChannel.
	bc js.Value
	// cb retains the registered JavaScript callback until Close.
	cb js.Func
}

// blockEventPub owns one send-only browser BroadcastChannel.
type blockEventPub struct {
	// bc is the underlying browser BroadcastChannel.
	bc js.Value
}

// main runs one cache probe scenario and reports its terminal result.
func main() {
	// Measure and parse one worker scenario from the browser launch arguments.
	started := time.Now()
	c, err := parseConfig(testArgs())

	// Run the selected storage action.
	if err == nil {
		err = run(context.Background(), c)
	}

	// Return one result before the WebAssembly runtime exits.
	postResult(c, time.Since(started), err)
}

// testArgs returns process arguments or the browser harness fallback.
func testArgs() []string {
	// Prefer the process arguments supplied by a standard WebAssembly launch.
	if len(os.Args) >= 7 {
		return os.Args
	}

	// Fall back to arguments installed by the browser worker harness.
	value := js.Global().Get("__OPFS_CHROMETEST_ARGS")
	if value.IsUndefined() || value.IsNull() {
		return os.Args
	}

	// Copy JavaScript values into Go-owned strings.
	args := make([]string, value.Get("length").Int())
	for i := range args {
		args[i] = value.Index(i).String()
	}
	return args
}

// parseConfig validates and converts the shared browser harness arguments.
func parseConfig(args []string) (*config, error) {
	// Require the complete shared worker argument shape.
	if len(args) < 7 {
		return nil, errors.Errorf("expected 6 args, got %d", len(args)-1)
	}

	// Parse every numeric workload dimension.
	worker, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.Wrap(err, "parse worker")
	}
	workers, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, errors.Wrap(err, "parse workers")
	}
	iterations, err := strconv.Atoi(args[5])
	if err != nil {
		return nil, errors.Wrap(err, "parse iterations")
	}
	batch, err := strconv.Atoi(args[6])
	if err != nil {
		return nil, errors.Wrap(err, "parse batch")
	}

	// Retain the scenario and root together with the validated dimensions.
	return &config{
		scenario:   args[1],
		root:       args[2],
		worker:     worker,
		workers:    workers,
		iterations: iterations,
		batch:      batch,
	}, nil
}

// run installs the remote OPFS driver and dispatches the selected scenario.
func run(ctx context.Context, c *config) error {
	// Install the browser harness driver before any OPFS access.
	opfs.InstallRemoteDriverFromGlobal()

	// Dispatch exactly one selected storage action.
	switch c.scenario {
	case "clear":
		return clearRoot(c.root)
	case "block-writer":
		return runBlockWriter(ctx, c)
	case "block-reader":
		return runBlockReader(ctx, c, false)
	case "block-reader-compact":
		return runBlockReader(ctx, c, true)
	case "block-verify":
		return runBlockVerify(ctx, c)
	default:
		return errors.Errorf("unknown cache probe scenario %q", c.scenario)
	}
}

// clearRoot recreates the shared test directory with no retained entries.
func clearRoot(rootName string) error {
	// Open the browser OPFS root.
	root, err := opfs.GetRoot()
	if err != nil {
		return err
	}

	// Reset and recreate the named shared-volume directory.
	err = opfs.DeleteEntry(root, rootName, true)
	if err != nil && !opfs.IsNotFound(err) {
		return err
	}
	_, err = opfs.GetDirectory(root, rootName, true)
	return err
}

// runBlockWriter publishes deterministic block batches and visibility events.
func runBlockWriter(ctx context.Context, c *config) error {
	// Open one publishing engine and its cross-runtime event channel.
	_, blocks, release, err := openBlockEngine(ctx, c)
	if err != nil {
		return err
	}
	defer release()

	// Open the publisher used to announce durable generations.
	events := newBlockEventPub(c.root)
	defer events.Close()

	// Publish deterministic batches and announce each visible generation.
	for i := range c.iterations {
		// Build one deterministic content-addressed batch.
		entries := make([]*block.PutBatchEntry, c.batch)
		for j := range entries {
			key := blockKey(c.worker, i, j)
			value := blockValue(key)
			ref, err := block.BuildBlockRef(value, nil)
			if err != nil {
				return errors.Wrap(err, "build concurrent block reference")
			}
			entries[j] = &block.PutBatchEntry{Ref: ref, Data: value}
		}

		// Publish and flush the complete generation.
		if err := blocks.PutBlockBatch(ctx, entries); err != nil {
			return errors.Wrap(err, "write concurrent blocks")
		}
		if _, err := blocks.Sync(ctx); err != nil {
			return errors.Wrap(err, "sync concurrent blocks")
		}

		// Announce the generation only after it is durable.
		events.Post(blockEvent{typ: "block-written", worker: c.worker, iteration: i})
	}

	// Announce completion after every published batch is durable.
	events.Post(blockEvent{typ: "block-writer-done", worker: c.worker})
	return nil
}

// runBlockReader observes concurrent writes and optionally verifies maintenance.
func runBlockReader(ctx context.Context, c *config, compact bool) error {
	// Open one reader before the publishers start.
	e, blocks, release, err := openBlockEngine(ctx, c)
	if err != nil {
		return err
	}
	defer release()

	// Subscribe before reporting readiness to the browser harness.
	events := newBlockEventSub(c)
	defer events.Close()

	// Announce subscription readiness before consuming publisher events.
	postReady(c)

	// Consume publication events until every writer completes.
	done := make([]bool, c.workers)
	var found int
	var doneCount int
	for doneCount < c.workers {
		// Wait for the next publication or completion event.
		event, err := events.Next(ctx)
		if err != nil {
			return err
		}
		switch event.typ {
		case "block-written":
			for j := range c.batch {
				// Resolve the expected content-addressed block.
				key := blockKey(event.worker, event.iteration, j)
				value, ok, err := getBlock(ctx, blocks, key)
				if err != nil {
					return errors.Wrap(err, "read concurrent block")
				}

				// A concurrent reader may observe the event before its next snapshot.
				if !ok {
					continue
				}

				// Count only blocks whose deterministic content matches.
				if !bytes.Equal(value, blockValue(key)) {
					return errors.Errorf("block value mismatch key=%s", string(key))
				}
				found++
			}
		case "block-writer-done":
			// Count each valid writer completion once.
			if event.worker < 0 || event.worker >= len(done) {
				return errors.Errorf("invalid writer id %d", event.worker)
			}
			if !done[event.worker] {
				done[event.worker] = true
				doneCount++
			}
		}
	}

	// Require evidence that the live reader observed concurrent data.
	if found == 0 {
		return errors.New("reader found no concurrently written blocks")
	}

	// Return after live reads unless this scenario also requests maintenance.
	if !compact {
		return nil
	}

	// Run bounded maintenance through the live reader before checking all values.
	if err := e.Maintenance(ctx); err != nil {
		return errors.Wrap(err, "maintain shared block volume")
	}
	return verifyBlocks(ctx, c, blocks, "after maintenance")
}

// runBlockVerify verifies all expected blocks after a fresh engine mount.
func runBlockVerify(ctx context.Context, c *config) error {
	// Open a fresh block adapter against the shared volume.
	_, blocks, release, err := openBlockEngine(ctx, c)
	if err != nil {
		return err
	}
	defer release()

	// Verify every expected block through the remounted adapter.
	return verifyBlocks(ctx, c, blocks, "after remount")
}

// verifyBlocks checks every deterministic key and value in the workload.
func verifyBlocks(ctx context.Context, c *config, blocks *engine.BlockStore, phase string) error {
	for w := range c.workers {
		for i := range c.iterations {
			for j := range c.batch {
				// Resolve one deterministic workload coordinate.
				key := blockKey(w, i, j)
				value, found, err := getBlock(ctx, blocks, key)
				if err != nil {
					return errors.Wrap(err, "read block "+phase)
				}

				// Require the block and its exact deterministic content.
				if !found {
					return errors.Errorf("missing block %s key=%s", phase, string(key))
				}
				if !bytes.Equal(value, blockValue(key)) {
					return errors.Errorf("bad block %s key=%s", phase, string(key))
				}
			}
		}
	}
	return nil
}

// openBlockEngine opens one runtime instance and its buffered block adapter.
func openBlockEngine(ctx context.Context, c *config) (*engine.Engine, *engine.BlockStore, func(), error) {
	// Open the shared test directory and its durable engine.
	dir, err := openTestDirectory(c.root, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	e, err := engine.Open(ctx, engine.NewBrowserBackend(opfs.DefaultDriver, dir, c.root))
	if err != nil {
		return nil, nil, nil, err
	}

	// Couple the block adapter and engine into one release operation.
	blocks := engine.NewBlockStore(ctx, e, block.DefaultHashType)
	release := func() {
		_ = blocks.Close()
		_ = e.Close()
	}
	return e, blocks, release, nil
}

// getBlock resolves the deterministic test payload to its content-addressed key.
func getBlock(ctx context.Context, blocks *engine.BlockStore, key []byte) ([]byte, bool, error) {
	ref, err := block.BuildBlockRef(blockValue(key), nil)
	if err != nil {
		return nil, false, err
	}
	return blocks.GetBlock(ctx, ref)
}

// openTestDirectory creates the named descendant path under the OPFS root.
func openTestDirectory(rootName string, parts []string) (js.Value, error) {
	// Open the browser OPFS root.
	root, err := opfs.GetRoot()
	if err != nil {
		return js.Undefined(), err
	}

	// Create the requested shared-volume path.
	path := append([]string{rootName}, parts...)
	return opfs.GetDirectoryPath(root, path, true)
}

// blockKey encodes one workload coordinate in lexical iteration order.
func blockKey(worker, iteration, entry int) []byte {
	return []byte("b/" + strconv.Itoa(worker) + "/" + zeroPad(iteration, 5) + "/" + zeroPad(entry, 3))
}

// blockValue derives deterministic test content from a workload key.
func blockValue(key []byte) []byte {
	return []byte("value:" + string(key))
}

// zeroPad renders a nonnegative workload index at the requested minimum width.
func zeroPad(n, width int) string {
	value := strconv.Itoa(n)
	for len(value) < width {
		value = "0" + value
	}
	return value
}

// newBlockEventSub subscribes to all events for one shared test root.
func newBlockEventSub(c *config) *blockEventSub {
	// Open the shared channel and buffer its complete event budget.
	ch := make(chan blockEvent, c.workers*c.iterations+c.workers+8)
	bc := js.Global().Get("BroadcastChannel").New(blockEventChannel(c.root))

	// Translate browser messages into the reader's Go event queue.
	cb := js.FuncOf(func(_ js.Value, args []js.Value) any {
		data := args[0].Get("data")
		ch <- blockEvent{
			typ:       data.Get("type").String(),
			worker:    data.Get("worker").Int(),
			iteration: data.Get("iteration").Int(),
		}
		return nil
	})

	// Retain the callback for the subscription lifetime.
	bc.Set("onmessage", cb)
	return &blockEventSub{ch: ch, bc: bc, cb: cb}
}

// Next waits for the next event or caller cancellation.
func (s *blockEventSub) Next(ctx context.Context) (blockEvent, error) {
	select {
	case event := <-s.ch:
		return event, nil
	case <-ctx.Done():
		return blockEvent{}, ctx.Err()
	}
}

// Close detaches the callback and releases browser resources.
func (s *blockEventSub) Close() {
	s.bc.Set("onmessage", js.Null())
	s.bc.Call("close")
	s.cb.Release()
}

// newBlockEventPub opens the send side of a shared test event channel.
func newBlockEventPub(root string) *blockEventPub {
	return &blockEventPub{bc: js.Global().Get("BroadcastChannel").New(blockEventChannel(root))}
}

// Post publishes one structured event to every reader.
func (p *blockEventPub) Post(event blockEvent) {
	// Encode the event as the browser harness message shape.
	obj := js.Global().Get("Object").New()
	obj.Set("type", event.typ)
	obj.Set("worker", event.worker)
	obj.Set("iteration", event.iteration)

	// Broadcast the event to every subscribed reader.
	p.bc.Call("postMessage", obj)
}

// Close releases the publisher's browser channel.
func (p *blockEventPub) Close() {
	p.bc.Call("close")
}

// blockEventChannel derives the BroadcastChannel name for one test root.
func blockEventChannel(root string) string {
	return "opfs-chrometest:" + root
}

// postReady announces that a reader subscribed before publishers may start.
func postReady(c *config) {
	// Encode the reader identity expected by the browser harness.
	obj := js.Global().Get("Object").New()
	obj.Set("kind", "ready")
	obj.Set("scenario", c.scenario)
	obj.Set("worker", c.worker)

	// Release the harness to start publishers.
	js.Global().Call("postMessage", obj)
}

// postResult reports one scenario's duration and terminal status.
func postResult(c *config, duration time.Duration, err error) {
	// Build the shared worker result fields.
	obj := js.Global().Get("Object").New()
	obj.Set("kind", "result")
	if c != nil {
		obj.Set("scenario", c.scenario)
		obj.Set("worker", c.worker)
	}

	// Attach the terminal status.
	obj.Set("durationMs", duration.Milliseconds())
	obj.Set("ok", true)
	if err != nil {
		obj.Set("ok", false)
		obj.Set("error", err.Error())
	}

	// Publish the terminal result to the browser harness.
	js.Global().Call("postMessage", obj)
}
