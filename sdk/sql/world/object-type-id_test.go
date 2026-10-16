package s4wave_sql_world_test

import (
	"testing"

	s4wave_sql_world "github.com/s4wave/spacewave/sdk/sql/world"
)

// TestSqlDbTypeIDResolvesWithoutFactory proves the SQL database type id
// resolves without the SRPC-server-backed factory. This file carries no build
// constraint, so the assertion compiles under TinyGo browser builds that
// exclude objecttype.go.
func TestSqlDbTypeIDResolvesWithoutFactory(t *testing.T) {
	if s4wave_sql_world.SqlDbTypeID != "sql/db" {
		t.Fatalf("unexpected SqlDbTypeID %q", s4wave_sql_world.SqlDbTypeID)
	}
}
