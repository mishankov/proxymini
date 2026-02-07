package requestlog

import (
	"net/http"
	"testing"
)

func TestNewSetsElapsedMS(t *testing.T) {
	t.Parallel()

	got := New(
		http.MethodPost,
		"http://proxy.local/api",
		"http://upstream.local/api",
		http.Header{"X-Test": []string{"1"}},
		`{"hello":"world"}`,
		http.StatusCreated,
		http.Header{"Content-Type": []string{"application/json"}},
		`{"ok":true}`,
		123,
	)

	if got.ElapsedMS != 123 {
		t.Fatalf("expected elapsed_ms=123, got %d", got.ElapsedMS)
	}
}
