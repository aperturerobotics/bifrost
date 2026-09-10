package sqlite_rpc

import (
	"context"
	"database/sql/driver"

	"github.com/pkg/errors"
	sql_sqlite_wasm_rpc "github.com/s4wave/spacewave/db/sql/sqlite-wasm/rpc"
)

// Connector binds one SQL pool to its database path and bridge lifetime.
type Connector struct {
	// driver resolves the bridge for each physical connection.
	driver *Driver
	// path identifies the database within that bridge.
	path string
}

// NewConnector constructs an instance-bound connector for sql.OpenDB.
func NewConnector(client sql_sqlite_wasm_rpc.SRPCSqliteBridgeClient, path string) *Connector {
	return &Connector{
		driver: NewDriver(func(context.Context) (sql_sqlite_wasm_rpc.SRPCSqliteBridgeClient, error) {
			return client, nil
		}),
		path: path,
	}
}

// Connect opens a physical connection using the pool's cancellation context.
func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	// Resolve the runtime's bridge before allocating a remote database handle.
	client, err := c.driver.resolve(ctx)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("sqlite-rpc: bridge is unavailable")
	}

	// Keep the connection bound to the bridge that allocated its handle.
	response, err := client.OpenDb(ctx, &sql_sqlite_wasm_rpc.OpenDbRequest{Path: c.path})
	if err != nil {
		return nil, errors.Wrap(err, "sqlite-rpc: open database")
	}
	return &connection{client: client, dbID: response.GetDbId()}, nil
}

// Driver returns the driver used to construct this connector.
func (c *Connector) Driver() driver.Driver {
	return c.driver
}

// _ is a type assertion.
var _ driver.Connector = (*Connector)(nil)
