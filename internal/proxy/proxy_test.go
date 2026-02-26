package proxy_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/db"
	"github.com/mishankov/proxymini/internal/proxy"
	"github.com/mishankov/proxymini/internal/requestlog"
)

func TestProxyRouting_PrefixMatching(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer("upstream-response", http.StatusOK)
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	handler := newTestProxyHandler(testDB, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body, _ := io.ReadAll(rr.Body)
	if string(body) != "upstream-response" {
		t.Errorf("expected body 'upstream-response', got '%s'", string(body))
	}
}

func TestProxyRouting_NoMatchingRoute(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	configContent := `[[proxy]]
prefix = "/api"
target = "http://localhost:9999"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	handler := newTestProxyHandler(testDB, conf)

	req := httptest.NewRequest(http.MethodGet, "/other/path", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), "no matching proxy found for URL:") {
		t.Errorf("expected error message to contain 'no matching proxy found for URL:', got '%s'", string(body))
	}
}

func TestProxyRequestHeaders_Forwarded(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	var receivedHeaders http.Header
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	handler := newTestProxyHandler(testDB, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("Authorization", "Bearer token123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if receivedHeaders.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("expected X-Custom-Header to be forwarded")
	}
	if receivedHeaders.Get("Authorization") != "Bearer token123" {
		t.Errorf("expected Authorization header to be forwarded")
	}
}

func TestProxyRequestBody_Preserved(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	var receivedBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	handler := newTestProxyHandler(testDB, conf)

	requestBody := `{"name":"test","value":123}`
	req := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if receivedBody != requestBody {
		t.Errorf("expected body '%s', got '%s'", requestBody, receivedBody)
	}
}

func TestProxyResponseHeaders_Preserved(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Response-Header", "response-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	handler := newTestProxyHandler(testDB, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Response-Header") != "response-value" {
		t.Errorf("expected X-Response-Header to be preserved")
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type to be preserved")
	}
}

func TestProxyResponseBody_Preserved(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	responseBody := `{"status":"success","id":"12345"}`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	handler := newTestProxyHandler(testDB, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Body)
	if string(body) != responseBody {
		t.Errorf("expected body '%s', got '%s'", responseBody, string(body))
	}
}

func TestProxySetsXProxyMiniHeader(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	upstream := newMockServer("response", http.StatusOK)
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	handler := newTestProxyHandler(testDB, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Proxy-Mini") != "true" {
		t.Errorf("expected X-Proxy-Mini header to be set to 'true'")
	}
}

func TestProxyTargetURLConstruction(t *testing.T) {
	testDB, cleanupDB := setupTestDB()
	defer cleanupDB()

	var receivedURL string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.Path
		if r.URL.RawQuery != "" {
			receivedURL += "?" + r.URL.RawQuery
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	configContent := `[[proxy]]
prefix = "/api"
target = "` + upstream.URL + `"`

	conf, cleanupConfig := createTestConfig(configContent)
	defer cleanupConfig()

	handler := newTestProxyHandler(testDB, conf)

	req := httptest.NewRequest(http.MethodGet, "/api/users?limit=10&offset=20", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	expectedURL := "/users?limit=10&offset=20"
	if receivedURL != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, receivedURL)
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

func newMockServer(responseBody string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

func newTestProxyHandler(testDB *sqlx.DB, conf *config.Config) *proxy.ProxyHandler {
	rlSvc := requestlog.NewRequestLogService(testDB)
	return proxy.NewProxyHandler(rlSvc, conf)
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
