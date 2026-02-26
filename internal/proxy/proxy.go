package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/requestlog"
	"github.com/mishankov/proxymini/internal/utils"
	"github.com/platforma-dev/platforma/log"
)

type ProxyHandler struct {
	rlSvc  *requestlog.RequestLogService
	config *config.Config
}

func NewProxyHandler(rlSvc *requestlog.RequestLogService, config *config.Config) *ProxyHandler {
	return &ProxyHandler{rlSvc: rlSvc, config: config}
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Proxy-Mini", "true")
	startedAt := time.Now()

	err := ph.config.ReloadProxies()
	if err != nil {
		handleError(w, fmt.Errorf("error getting config: %w", err), http.StatusInternalServerError)
		return
	}

	target := ""
	prefix := ""
	skipLogging := false
	for _, proxy := range ph.config.Proxies {
		if strings.HasPrefix(r.URL.Path, proxy.Prefix) {
			target = proxy.Target
			prefix = proxy.Prefix
			skipLogging = proxy.SkipLogging
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

	targetUrl := target + strings.TrimPrefix(r.URL.Path, prefix)
	if r.URL.RawQuery != "" {
		targetUrl += "?" + r.URL.RawQuery
	}
	if r.URL.Fragment != "" {
		targetUrl += "#" + r.URL.Fragment
	}

	req, err := http.NewRequest(r.Method, targetUrl, bytes.NewReader(reqBody))
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

	elapsedMS := time.Since(startedAt).Milliseconds()

	if !skipLogging {
		reqLog := requestlog.New(
			r.Method,
			fullURL(r),
			targetUrl,
			r.Header,
			string(reqBody),
			resp.StatusCode,
			resp.Header,
			string(body),
			elapsedMS,
		)

		err = ph.rlSvc.Save(reqLog)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to save request log", "error", err)
		}
	}
}

func handleError(w http.ResponseWriter, err error, status int) {
	log.ErrorContext(context.Background(), "request handling error", "status", status, "error", err)
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}

// fullURL returns the full URL of the incoming request, including protocol, host, path, query parameters, and fragment.
func fullURL(r *http.Request) string {
	builder := strings.Builder{}

	if r.TLS != nil {
		builder.WriteString("https://")
	} else {
		builder.WriteString("http://")
	}

	builder.WriteString(r.Host)
	builder.WriteString(r.URL.Path)

	if r.URL.RawQuery != "" {
		builder.WriteString("?" + r.URL.RawQuery)
	}

	if r.URL.Fragment != "" {
		builder.WriteString("#" + r.URL.Fragment)
	}

	return builder.String()
}
