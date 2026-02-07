package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/db"
	"github.com/mishankov/proxymini/internal/requestlog"
	"github.com/mishankov/proxymini/internal/services"
)

func TestProxyHandlerStoresElapsedMSAndReturnsViaAPI(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Millisecond)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "proxy.db")
	confPath := filepath.Join(tempDir, "proxymini.conf.toml")
	confContent := "[[proxy]]\nprefix = \"/api\"\ntarget = \"" + upstream.URL + "\"\n"
	if err := os.WriteFile(confPath, []byte(confContent), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	conn, err := db.Connect(dbPath)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer conn.Close()

	if err := db.Init(conn); err != nil {
		t.Fatalf("db init failed: %v", err)
	}

	rls := services.NewRequestLogService(conn)
	ph := NewProxyHandler(rls, &config.Config{ConfigPath: confPath})
	rlh := NewRequestLogHandler(rls)

	mux := http.NewServeMux()
	mux.Handle("/api/logs", rlh)
	mux.Handle("/", ph)

	proxyReq := httptest.NewRequest(http.MethodGet, "http://proxy.local/api/test?x=1", nil)
	proxyRec := httptest.NewRecorder()
	mux.ServeHTTP(proxyRec, proxyReq)

	if proxyRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, proxyRec.Code)
	}

	var payload []requestlog.RequestLog
	deadline := time.Now().Add(2 * time.Second)
	for {
		logReq := httptest.NewRequest(http.MethodGet, "http://proxy.local/api/logs", nil)
		logRec := httptest.NewRecorder()
		mux.ServeHTTP(logRec, logReq)

		if logRec.Code != http.StatusOK {
			t.Fatalf("logs endpoint returned status %d", logRec.Code)
		}

		if err := json.Unmarshal(logRec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode logs response failed: %v", err)
		}

		if len(payload) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for async request log save")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if payload[0].ElapsedMS < 0 {
		t.Fatalf("expected elapsed_ms >= 0, got %d", payload[0].ElapsedMS)
	}
}
