package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"

	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/requestlog"
	"github.com/mishankov/proxymini/internal/services"
	"github.com/mishankov/proxymini/internal/utils"
)

type ProxyHandler struct {
	rlSvc *services.RequestLogService
}

func NewProxyHandler(rlSvc *services.RequestLogService) *ProxyHandler {
	return &ProxyHandler{rlSvc: rlSvc}
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Proxy-Mini", "true")

	config, err := config.New()
	if err != nil {
		handleError(w, fmt.Errorf("error getting config: %w", err), http.StatusInternalServerError)
		return
	}

	target := ""
	prefix := ""
	for _, proxy := range config.Proxies {
		if strings.HasPrefix(r.URL.Path, proxy.Prefix) {
			target = proxy.Target
			prefix = proxy.Prefix
		}
	}

	if target == "" {
		handleError(w, fmt.Errorf("no matching proxy found for URL: %s", fullURL(r)), http.StatusNotFound)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, fmt.Errorf("error reading request body: %w", err), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(r.Method, target+strings.TrimPrefix(r.URL.Path, prefix), bytes.NewReader(reqBody))
	if err != nil {
		handleError(w, fmt.Errorf("error creating request: %w", err), http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		handleError(w, fmt.Errorf("error making request: %w", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for hn, hvs := range resp.Header {
		for _, hv := range hvs {
			w.Header().Add(hn, hv)
		}
	}

	http.NewResponseController(w).Flush()

	body, err := utils.CopyBuffer(w, resp.Body, []byte{})
	// Some clients cause `write: broken pipe` error in the end of a request. This seems to be ok, so ignore `syscall.EPIPE`.
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		handleError(w, fmt.Errorf("error copying buffer: %w", err), http.StatusInternalServerError)
	}

	reqLog := requestlog.New(r.Method, fullURL(r), r.Header, string(reqBody), resp.StatusCode, resp.Header, string(body))

	ph.rlSvc.Save(reqLog)
}
