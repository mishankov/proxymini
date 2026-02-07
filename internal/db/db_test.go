package db

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func hasColumn(t *testing.T, db *sqlx.DB, table, column string) bool {
	t.Helper()

	rows, err := db.Queryx("PRAGMA table_info(" + table + ")")
	if err != nil {
		t.Fatalf("pragma table_info failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notnull int
		var dfltValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan pragma row failed: %v", err)
		}
		if name == column {
			return true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate pragma rows failed: %v", err)
	}

	return false
}

func TestInitCreatesElapsedMSColumnOnFreshDB(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "fresh.db")
	conn, err := Connect(dbPath)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer conn.Close()

	if err := Init(conn); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if !hasColumn(t, conn, "request_log", "elapsed_ms") {
		t.Fatalf("expected elapsed_ms column to exist")
	}
}

func TestInitMigratesLegacyDBWithoutElapsedMS(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	conn, err := Connect(dbPath)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer conn.Close()

	_, err = conn.Exec(`
CREATE TABLE request_log (
    id TEXT PRIMARY KEY,
    time BIGINT NOT NULL,
    method TEXT,
    proxy_url TEXT,
    url TEXT,
    request_headers TEXT,
    request_body TEXT,
    status INT NOT NULL,
    response_headers TEXT,
    response_body TEXT
);`)
	if err != nil {
		t.Fatalf("create legacy table failed: %v", err)
	}

	if hasColumn(t, conn, "request_log", "elapsed_ms") {
		t.Fatalf("expected legacy schema to not contain elapsed_ms before migration")
	}

	if err := Init(conn); err != nil {
		t.Fatalf("init migration failed: %v", err)
	}

	if !hasColumn(t, conn, "request_log", "elapsed_ms") {
		t.Fatalf("expected elapsed_ms column to exist after migration")
	}
}
