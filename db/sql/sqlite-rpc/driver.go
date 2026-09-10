package sqlite_rpc

import (
	"context"
	"database/sql/driver"

	sql_sqlite_wasm_rpc "github.com/s4wave/spacewave/db/sql/sqlite-wasm/rpc"
)

// ClientResolver waits for the SQLite bridge that will own a connection.
type ClientResolver func(context.Context) (sql_sqlite_wasm_rpc.SRPCSqliteBridgeClient, error)

// Driver connects SQL pools to the bridge supplied by their runtime.
type Driver struct {
	// resolve obtains the bridge when database/sql opens a physical connection.
	resolve ClientResolver
}

// NewDriver constructs a driver whose connector retains its runtime resolver.
func NewDriver(resolve ClientResolver) *Driver {
	return &Driver{resolve: resolve}
}

// Open opens a connection for callers that do not support DriverContext.
func (d *Driver) Open(path string) (driver.Conn, error) {
	connector, err := d.OpenConnector(path)
	if err != nil {
		return nil, err
	}
	return connector.Connect(context.Background())
}

// OpenConnector binds a database path without opening a connection.
func (d *Driver) OpenConnector(path string) (driver.Connector, error) {
	return &Connector{driver: d, path: path}, nil
}

// _ is a type assertion.
var _ driver.DriverContext = (*Driver)(nil)
