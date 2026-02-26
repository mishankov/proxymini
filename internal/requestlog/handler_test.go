package requestlog_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mishankov/proxymini/internal/db"
	"github.com/mishankov/proxymini/internal/requestlog"
)

func TestGetLogs_EmptyDB_ReturnsEmptyArray(t *testing.T) {
	testDB, cleanup := setupTestDB()
	defer cleanup()

	handler := newTestRequestLogHandler(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var logs []requestlog.RequestLog
	err := json.Unmarshal(rr.Body.Bytes(), &logs)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("expected empty array, got %d logs", len(logs))
	}
}

func TestGetLogs_WithLogs_ReturnsCorrectData(t *testing.T) {
	testDB, cleanup := setupTestDB()
	defer cleanup()

	rlSvc := requestlog.NewRequestLogService(testDB)

	log1 := createTestRequestLog("GET", "http://example.com/api/1")
	log2 := createTestRequestLog("POST", "http://example.com/api/2")

	rlSvc.Save(log1)
	time.Sleep(10 * time.Millisecond)
	rlSvc.Save(log2)

	time.Sleep(50 * time.Millisecond)

	handler := newTestRequestLogHandler(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var logs []requestlog.RequestLog
	err := json.Unmarshal(rr.Body.Bytes(), &logs)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}

	methods := map[string]bool{}
	for _, log := range logs {
		methods[log.Method] = true
	}

	if !methods["GET"] {
		t.Errorf("expected to find GET method in logs")
	}

	if !methods["POST"] {
		t.Errorf("expected to find POST method in logs")
	}
}

func TestGetLogs_ReturnsLogsInDescendingOrder(t *testing.T) {
	testDB, cleanup := setupTestDB()
	defer cleanup()

	rlSvc := requestlog.NewRequestLogService(testDB)

	for i := 0; i < 3; i++ {
		log := createTestRequestLog("GET", "http://example.com/api")
		rlSvc.Save(log)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)

	handler := newTestRequestLogHandler(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var logs []requestlog.RequestLog
	err := json.Unmarshal(rr.Body.Bytes(), &logs)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(logs))
	}

	for i := 1; i < len(logs); i++ {
		if logs[i-1].Time < logs[i].Time {
			t.Errorf("logs not in descending order: log[%d].Time=%d < log[%d].Time=%d", i-1, logs[i-1].Time, i, logs[i].Time)
		}
	}
}

func TestDeleteLogs_RemovesAllLogs(t *testing.T) {
	testDB, cleanup := setupTestDB()
	defer cleanup()

	rlSvc := requestlog.NewRequestLogService(testDB)

	log := createTestRequestLog("GET", "http://example.com/api")
	rlSvc.Save(log)

	time.Sleep(50 * time.Millisecond)

	handler := newTestRequestLogHandler(testDB)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/logs", nil)
	deleteRr := httptest.NewRecorder()

	handler.ServeHTTP(deleteRr, deleteReq)

	if deleteRr.Code != http.StatusOK {
		t.Errorf("expected status %d for DELETE, got %d", http.StatusOK, deleteRr.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	getRr := httptest.NewRecorder()

	handler.ServeHTTP(getRr, getReq)

	var logs []requestlog.RequestLog
	err := json.Unmarshal(getRr.Body.Bytes(), &logs)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("expected empty array after delete, got %d logs", len(logs))
	}
}

func TestGetLogs_JSONSerialization(t *testing.T) {
	testDB, cleanup := setupTestDB()
	defer cleanup()

	rlSvc := requestlog.NewRequestLogService(testDB)

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	log := requestlog.New(
		"POST",
		"http://proxy.example.com/api",
		"http://upstream.example.com/api",
		headers,
		`{"key":"value"}`,
		http.StatusCreated,
		headers,
		`{"result":"ok"}`,
		42,
	)

	rlSvc.Save(log)

	time.Sleep(50 * time.Millisecond)

	handler := newTestRequestLogHandler(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var logs []requestlog.RequestLog
	err := json.Unmarshal(rr.Body.Bytes(), &logs)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}

	if logs[0].Method != "POST" {
		t.Errorf("expected Method 'POST', got '%s'", logs[0].Method)
	}

	if logs[0].ProxyURL != "http://proxy.example.com/api" {
		t.Errorf("expected ProxyURL 'http://proxy.example.com/api', got '%s'", logs[0].ProxyURL)
	}

	if logs[0].URL != "http://upstream.example.com/api" {
		t.Errorf("expected URL 'http://upstream.example.com/api', got '%s'", logs[0].URL)
	}

	if logs[0].Status != http.StatusCreated {
		t.Errorf("expected Status %d, got %d", http.StatusCreated, logs[0].Status)
	}

	if logs[0].ElapsedMS != 42 {
		t.Errorf("expected ElapsedMS 42, got %d", logs[0].ElapsedMS)
	}

	if logs[0].RequestBody != `{"key":"value"}` {
		t.Errorf("expected RequestBody '{\"key\":\"value\"}', got '%s'", logs[0].RequestBody)
	}

	if logs[0].ResponseBody != `{"result":"ok"}` {
		t.Errorf("expected ResponseBody '{\"result\":\"ok\"}', got '%s'", logs[0].ResponseBody)
	}
}

func TestDeleteOlderThan_RemovesOnlyOldLogs(t *testing.T) {
	testDB, cleanup := setupTestDB()
	defer cleanup()

	rlSvc := requestlog.NewRequestLogService(testDB)

	// Create logs with specific timestamps
	now := time.Now().UTC().Unix()
	oldLog := requestlog.RequestLog{
		ID:       "old-log-id",
		Time:     now - 100, // 100 seconds ago
		Method:   "GET",
		ProxyURL: "http://example.com/old",
		Status:   200,
	}
	newLog := requestlog.RequestLog{
		ID:       "new-log-id",
		Time:     now - 10, // 10 seconds ago
		Method:   "POST",
		ProxyURL: "http://example.com/new",
		Status:   201,
	}

	rlSvc.Save(oldLog)
	rlSvc.Save(newLog)

	time.Sleep(50 * time.Millisecond)

	// Delete logs older than 50 seconds
	threshold := now - 50
	err := rlSvc.DeleteOlderThan(threshold)
	if err != nil {
		t.Fatalf("failed to delete old logs: %v", err)
	}

	// Verify only new log remains
	logs, err := rlSvc.GetList()
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("expected 1 log after deletion, got %d", len(logs))
	}

	if len(logs) > 0 && logs[0].ID != "new-log-id" {
		t.Errorf("expected remaining log to be 'new-log-id', got '%s'", logs[0].ID)
	}
}

func createTestRequestLog(method, url string) requestlog.RequestLog {
	return requestlog.New(
		method,
		url,
		"http://upstream.example.com",
		http.Header{},
		"",
		http.StatusOK,
		http.Header{},
		"",
		10,
	)
}

func setupTestDB() (*sqlx.DB, func()) {
	testDB, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}

	err = db.Init(testDB)
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		testDB.Close()
	}

	return testDB, cleanup
}

func newTestRequestLogHandler(testDB *sqlx.DB) *requestlog.RequestLogHandler {
	rlSvc := requestlog.NewRequestLogService(testDB)
	return requestlog.NewRequestLogHandler(rlSvc)
}
