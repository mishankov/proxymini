package services

import (
	"path/filepath"
	"testing"

	"github.com/mishankov/proxymini/internal/db"
	"github.com/mishankov/proxymini/internal/requestlog"
)

func TestSaveAndGetListPersistElapsedMS(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "service.db")
	conn, err := db.Connect(dbPath)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer conn.Close()

	if err := db.Init(conn); err != nil {
		t.Fatalf("db init failed: %v", err)
	}

	rls := NewRequestLogService(conn)

	entry := requestlog.RequestLog{
		ID:              "test-id",
		Time:            100,
		ElapsedMS:       321,
		Method:          "GET",
		ProxyURL:        "http://proxy.local/api",
		URL:             "http://upstream.local/api",
		RequestHeaders:  `{"Accept":["*/*"]}`,
		RequestBody:     "",
		Status:          200,
		ResponseHeaders: `{"Content-Type":["application/json"]}`,
		ResponseBody:    `{"ok":true}`,
	}

	if err := rls.save(entry); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := rls.GetList()
	if err != nil {
		t.Fatalf("get list failed: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got))
	}
	if got[0].ElapsedMS != entry.ElapsedMS {
		t.Fatalf("expected elapsed_ms=%d, got %d", entry.ElapsedMS, got[0].ElapsedMS)
	}
}
