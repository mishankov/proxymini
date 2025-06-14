package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"syscall"

	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/requestlog"
	"github.com/mishankov/proxymini/internal/services"
	"github.com/mishankov/proxymini/internal/utils"
)

type ProxyHandler struct {
	rlSvc  *services.RequestLogService
	config *config.Config
}

func NewProxyHandler(rlSvc *services.RequestLogService, config *config.Config) *ProxyHandler {
	return &ProxyHandler{rlSvc: rlSvc, config: config}
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Proxy-Mini", "true")

	err := ph.config.ReloadProxies()
	if err != nil {
		handleError(w, fmt.Errorf("error getting config: %w", err), http.StatusInternalServerError)
		return
	}

	target := ""
	prefix := ""
	for _, proxy := range ph.config.Proxies {
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

	for hn, hvs := range r.Header {
		for _, hv := range hvs {
			req.Header.Add(hn, hv)
		}
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

	w.WriteHeader(resp.StatusCode)

	body, err := utils.CopyBuffer(w, resp.Body, []byte{})
	// Some clients cause `write: broken pipe` error in the end of a request. This seems to be ok, so ignore `syscall.EPIPE`.
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		handleError(w, fmt.Errorf("error copying buffer: %w", err), http.StatusInternalServerError)
	}

	reqLog := requestlog.New(r.Method, fullURL(r), r.Header, string(reqBody), resp.StatusCode, resp.Header, string(body))

	err = ph.rlSvc.Save(reqLog)
	if err != nil {
		log.Println("Error saving log:", err)
	}
}
