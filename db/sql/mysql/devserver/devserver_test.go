package mysql_devserver

import (
	"context"
	"database/sql"
	"net"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// TestServeSchema builds a database from a migration directory and reads it
// back over the MySQL wire protocol.
func TestServeSchema(t *testing.T) {
	// Write two migrations, the second without a final semicolon, and a
	// file the directory read skips.
	ctx := context.Background()
	dir := t.TempDir()
	for name, text := range map[string]string{
		"0001_users.sql": "create table users (id int primary key, name text not null);\n",
		"0002_posts.sql": "create table posts (id int primary key, user_id int not null)",
		"README.md":      "not sql",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	files, texts, err := ReadSchemaFiles([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || filepath.Base(files[0]) != "0001_users.sql" {
		t.Fatalf("schema files=%v want the two migrations in order", files)
	}

	// Serve the schema and connect.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := Start(ctx, lis, "app", texts)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	db, err := sql.Open("mysql", "root@tcp("+lis.Addr().String()+")/app")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// A prepared join over both tables answers with no rows.
	rows, err := db.QueryContext(ctx, "select u.name from users u join posts p on p.user_id = u.id where u.id = ?", 1)
	if err != nil {
		t.Fatal(err)
	}
	cols, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	if rows.Next() {
		t.Fatal("want no rows")
	}
	_ = rows.Close()
	if len(cols) != 1 || cols[0] != "name" {
		t.Fatalf("columns=%v want [name]", cols)
	}

	// The database is read-only.
	if _, err := db.ExecContext(ctx, "insert into users (id, name) values (1, 'a')"); err == nil {
		t.Fatal("want insert to fail")
	}
}
