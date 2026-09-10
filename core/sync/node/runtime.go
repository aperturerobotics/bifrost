//go:build js

// Package core_sync_node connects the compiled World to an instance's SQLite and Resource ports.
package core_sync_node

import (
	"context"
	"errors"
	"sync"
	"syscall/js"

	"github.com/aperturerobotics/starpc/srpc"
	storage_node "github.com/s4wave/spacewave/bldr/storage/node"
	message_port "github.com/s4wave/spacewave/bldr/web/entrypoint/browser/message-port"
	core_sync "github.com/s4wave/spacewave/core/sync"
	sqlite_bridge "github.com/s4wave/spacewave/db/sql/sqlite-wasm/rpc"
	"github.com/sirupsen/logrus"
)

// Runtime owns the engine and its private Resource and SQL stream lifetimes.
type Runtime struct {
	// engine owns the controller bus and ready World.
	engine *core_sync.Engine
	// mux serves the engine through the existing Resource protocol.
	mux srpc.Invoker
	// releaseMount releases resources without deleting the engine's data.
	releaseMount func()
	// ctx and cancel bound SQL activity, including calls with background contexts.
	ctx    context.Context
	cancel context.CancelFunc
	// serverCtx and stopServer terminate admission and active application requests first.
	serverCtx  context.Context
	stopServer context.CancelFunc
	// mtx serializes stream admission with shutdown.
	mtx sync.Mutex
	// closing rejects new Resource streams while storage is being fenced.
	closing bool
	// serverStreams joins every accepted Resource stream.
	serverStreams sync.WaitGroup
	// sqlStreams joins every local SQL bridge stream.
	sqlStreams sync.WaitGroup
	// closeOnce and closeErr retain one joined shutdown result.
	closeOnce sync.Once
	closeErr  error
}

// Open starts one engine using the caller's synchronous SQL-port opener.
// Each returned port carries existing StaRPC packets, not application controls.
// The caller holds exclusive directory ownership until Close has completed.
func Open(directory string, openSQLPort js.Value) (*Runtime, error) {
	ctx, cancel := context.WithCancel(context.Background())
	serverCtx, stopServer := context.WithCancel(ctx)
	r := &Runtime{ctx: ctx, cancel: cancel, serverCtx: serverCtx, stopServer: stopServer}
	le := logrus.NewEntry(logrus.New())
	client := srpc.NewClient(func(ctx context.Context, packet srpc.PacketDataHandler, closed srpc.CloseHandler) (srpc.PacketWriter, error) {
		r.mtx.Lock()
		if r.ctx.Err() != nil {
			r.mtx.Unlock()
			return nil, r.ctx.Err()
		}
		r.sqlStreams.Add(1)
		r.mtx.Unlock()
		port := message_port.NewMessagePort(openSQLPort.Invoke())
		stream := message_port.NewMessagePortPacketStream(port)
		ctx, cancel := context.WithCancel(ctx)
		stop := context.AfterFunc(r.ctx, cancel)
		go func() {
			defer r.sqlStreams.Done()
			defer cancel()
			defer stop()
			defer port.Close()
			stream.ReadPump(ctx, packet, closed)
		}()
		return stream, nil
	})
	engine, err := core_sync.Open(ctx, le, storage_node.NewStorage(sqlite_bridge.NewSRPCSqliteBridgeClient(client), directory))
	if err != nil {
		return nil, errors.Join(err, r.Close())
	}
	r.engine = engine
	r.mux, r.releaseMount, err = core_sync.AttachEngine(le, engine.Bus, engine.World)
	if err != nil {
		return nil, errors.Join(err, r.Close())
	}
	return r, nil
}

// Accept serves one Resource RPC on a caller-supplied packet port.
func (r *Runtime) Accept(value js.Value) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.closing {
		return errors.New("sync engine is closing")
	}
	port := message_port.NewMessagePort(value)
	stream := message_port.NewMessagePortPacketStream(port)
	server := srpc.NewServerRPC(r.serverCtx, r.mux, stream)
	r.serverStreams.Add(1)
	go func() {
		defer r.serverStreams.Done()
		defer port.Close()
		stream.ReadPump(r.serverCtx, server.HandlePacketData, server.HandleStreamClose)
	}()
	return nil
}

// Close stops requests, fences and closes the World, then releases SQL streams.
// The caller can release the directory lock after this method returns.
func (r *Runtime) Close() error {
	r.closeOnce.Do(func() {
		r.mtx.Lock()
		r.closing = true
		r.stopServer()
		r.mtx.Unlock()
		r.serverStreams.Wait()
		if r.releaseMount != nil {
			r.releaseMount()
		}
		if r.engine != nil {
			r.closeErr = r.engine.Close()
		}
		r.mtx.Lock()
		r.cancel()
		r.mtx.Unlock()
		r.sqlStreams.Wait()
	})
	return r.closeErr
}
