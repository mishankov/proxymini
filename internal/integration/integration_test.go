package tests_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/db"
	"github.com/mishankov/proxymini/internal/proxy"
	"github.com/mishankov/proxymini/internal/requestlog"
)

func TestProxyRequest_LogsSaved(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer(`{"status":"ok"}`, http.StatusOK, http.Header{
		"Content-Type": []string{"application/json"},
	})
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	rlSvc := requestlog.NewRequestLogService(testDB)
	proxyHandler := proxy.NewProxyHandler(rlSvc, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	proxyHandler.ServeHTTP(rr, req)

	time.Sleep(50 * time.Millisecond)

	logs, err := rlSvc.GetList()
	if err == nil {
		t.Fatalf("failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

func TestProxyRequest_LogsContainCorrectData(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer(`{"result":"success"}`, http.StatusCreated, http.Header{
		"X-Custom-Response": []string{"custom-value"},
	})
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	rlSvc := requestlog.NewRequestLogService(testDB)
	proxyHandler := proxy.NewProxyHandler(rlSvc, conf)

	requestBody := `{"name":"test user","action":"create"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users?org=acme", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()

	proxyHandler.ServeHTTP(rr, req)

	time.Sleep(50 * time.Millisecond)

	logs, err := rlSvc.GetList()
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}

	log := logs[0]

	if log.Method != "POST" {
		t.Errorf("expected Method 'POST', got '%s'", log.Method)
	}

	if !strings.Contains(log.ProxyURL, "/api/users") {
		t.Errorf("expected ProxyURL to contain '/api/users', got '%s'", log.ProxyURL)
	}

	if !strings.Contains(log.ProxyURL, "org=acme") {
		t.Errorf("expected ProxyURL to contain query params, got '%s'", log.ProxyURL)
	}

	if log.RequestBody != requestBody {
		t.Errorf("expected RequestBody '%s', got '%s'", requestBody, log.RequestBody)
	}

	if log.Status != http.StatusCreated {
		t.Errorf("expected Status %d, got %d", http.StatusCreated, log.Status)
	}

	if log.ResponseBody != `{"result":"success"}` {
		t.Errorf("expected ResponseBody '{\"result\":\"success\"}', got '%s'", log.ResponseBody)
	}

	if log.ElapsedMS < 0 {
		t.Errorf("expected ElapsedMS >= 0, got %d", log.ElapsedMS)
	}

	if log.Time <= 0 {
		t.Errorf("expected Time > 0, got %d", log.Time)
	}

	if log.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestProxyRequest_LogsContainRequestHeaders(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer("ok", http.StatusOK, nil)
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	rlSvc := requestlog.NewRequestLogService(testDB)
	proxyHandler := proxy.NewProxyHandler(rlSvc, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()

	proxyHandler.ServeHTTP(rr, req)

	time.Sleep(50 * time.Millisecond)

	logs, err := rlSvc.GetList()
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}

	var requestHeaders http.Header
	err = json.Unmarshal([]byte(logs[0].RequestHeaders), &requestHeaders)
	if err != nil {
		t.Fatalf("failed to unmarshal request headers: %v", err)
	}

	if requestHeaders.Get("X-Request-ID") != "req-123" {
		t.Errorf("expected X-Request-ID 'req-123', got '%s'", requestHeaders.Get("X-Request-ID"))
	}
}

func TestProxyRequest_LogsContainResponseHeaders(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer("ok", http.StatusOK, http.Header{
		"X-Response-ID": []string{"resp-456"},
		"Content-Type":  []string{"text/plain"},
	})
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	rlSvc := requestlog.NewRequestLogService(testDB)
	proxyHandler := proxy.NewProxyHandler(rlSvc, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	proxyHandler.ServeHTTP(rr, req)

	time.Sleep(50 * time.Millisecond)

	logs, err := rlSvc.GetList()
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}

	var responseHeaders http.Header
	err = json.Unmarshal([]byte(logs[0].ResponseHeaders), &responseHeaders)
	if err != nil {
		t.Fatalf("failed to unmarshal response headers: %v", err)
	}

	if responseHeaders.Get("X-Response-ID") != "resp-456" {
		t.Errorf("expected X-Response-ID 'resp-456', got '%s'", responseHeaders.Get("X-Response-ID"))
	}
}

func TestProxyRequest_SkipLoggingPreventsLogCreation(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer("ok", http.StatusOK, nil)
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/health"
target = "` + upstream.URL + `"
skipLogging = true

[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	rlSvc := requestlog.NewRequestLogService(testDB)
	proxyHandler := proxy.NewProxyHandler(rlSvc, conf)

	req1 := httptest.NewRequest(http.MethodGet, "/health/check", nil)
	rr1 := httptest.NewRecorder()
	proxyHandler.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr2 := httptest.NewRecorder()
	proxyHandler.ServeHTTP(rr2, req2)

	time.Sleep(50 * time.Millisecond)

	logs, err := rlSvc.GetList()
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 log (only /api request), got %d", len(logs))
	}

	if logs[0].Method != "GET" {
		t.Errorf("expected logged request to be GET, got %s", logs[0].Method)
	}

	if !strings.Contains(logs[0].ProxyURL, "/api") {
		t.Errorf("expected logged request to be /api, got %s", logs[0].ProxyURL)
	}
}

func TestProxyRequest_MultipleRequests_MultipleLogs(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer("ok", http.StatusOK, nil)
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	rlSvc := requestlog.NewRequestLogService(testDB)
	proxyHandler := proxy.NewProxyHandler(rlSvc, conf)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
		rr := httptest.NewRecorder()
		proxyHandler.ServeHTTP(rr, req)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)

	logs, err := rlSvc.GetList()
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}

	if len(logs) != 5 {
		t.Errorf("expected 5 logs, got %d", len(logs))
	}
}

func TestFullFlow_ViaHTTPHandlers(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer(`{"data":"test"}`, http.StatusOK, http.Header{
		"Content-Type": []string{"application/json"},
	})
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	rlSvc := requestlog.NewRequestLogService(testDB)
	rlHandler := requestlog.NewRequestLogHandler(rlSvc)
	proxyHandler := proxy.NewProxyHandler(rlSvc, conf)

	mux := http.NewServeMux()
	mux.Handle("/api/logs", rlHandler)
	mux.Handle("/", proxyHandler)
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/test")
	if err != nil {
		t.Fatalf("failed to make proxy request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"data":"test"}` {
		t.Errorf("expected body '{\"data\":\"test\"}', got '%s'", string(body))
	}

	time.Sleep(50 * time.Millisecond)

	logsResp, err := http.Get(server.URL + "/api/logs")
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}
	defer logsResp.Body.Close()

	if logsResp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d for logs endpoint, got %d", http.StatusOK, logsResp.StatusCode)
	}

	var logs []requestlog.RequestLog
	err = json.NewDecoder(logsResp.Body).Decode(&logs)
	if err != nil {
		t.Fatalf("failed to decode logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}

	if logs[0].Status != http.StatusOK {
		t.Errorf("expected logged status %d, got %d", http.StatusOK, logs[0].Status)
	}
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

func newMockServer(responseBody string, statusCode int, headers http.Header) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, values := range headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

func createTestConfig(content string) (*config.Config, func()) {
	tmpFile, err := os.CreateTemp("", "proxymini-test-*.toml")
	if err != nil {
		panic(err)
	}

	_, err = tmpFile.WriteString(content)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		panic(err)
	}
	tmpFile.Close()

	conf := &config.Config{
		Port:       "14443",
		ConfigPath: tmpFile.Name(),
		DBPath:     ":memory:",
	}

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return conf, cleanup
}
