//go:build js

// Package sqlite_wasm selects the browser SQLite Worker for the shared RPC driver.
package sqlite_wasm

import (
	"context"
	"database/sql"

	"github.com/aperturerobotics/util/broadcast"
	"github.com/pkg/errors"
	sqlite_rpc "github.com/s4wave/spacewave/db/sql/sqlite-rpc"
	sql_sqlite_wasm_rpc "github.com/s4wave/spacewave/db/sql/sqlite-wasm/rpc"
)

// driverName identifies the browser's readiness-bound SQL driver.
const driverName = "sqlite3-wasm"

// clientBcast guards globalClient and wakes getClient waiters on every change.
var clientBcast broadcast.Broadcast

// globalClient is the current RPC client, set by SetClient.
var globalClient sql_sqlite_wasm_rpc.SRPCSqliteBridgeClient

func init() {
	sql.Register(driverName, sqlite_rpc.NewDriver(getClient))
}

// SetClient sets the RPC client used by the driver.
// Call with nil to clear the client (e.g. on Worker disconnect).
func SetClient(client sql_sqlite_wasm_rpc.SRPCSqliteBridgeClient) {
	clientBcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		globalClient = client
		broadcast()
	})
}

// getClient returns the current RPC client, blocking until one is available.
func getClient(ctx context.Context) (sql_sqlite_wasm_rpc.SRPCSqliteBridgeClient, error) {
	for {
		var c sql_sqlite_wasm_rpc.SRPCSqliteBridgeClient
		var waitCh <-chan struct{}
		clientBcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			c = globalClient
			waitCh = getWaitCh()
		})
		if c != nil {
			return c, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-waitCh:
		}
	}
}

// DeleteDatabase deletes a database by path via the RPC client.
func DeleteDatabase(path string) error {
	client, err := getClient(context.Background())
	if err != nil {
		return errors.Wrap(err, "sqlite-wasm: get client for delete")
	}
	_, err = client.DeleteDb(context.Background(), &sql_sqlite_wasm_rpc.DeleteDbRequest{Path: path})
	return err
}
