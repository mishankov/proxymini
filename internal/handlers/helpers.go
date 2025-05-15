package handlers

import (
	"log"
	"net/http"
	"strings"
)

// handleError writes an error response to the client with the specified status code.
// It logs the error for debugging purposes and sends the error message as the response body.
func handleError(w http.ResponseWriter, err error, status int) {
	log.Println(err)
	// w.WriteHeader(status)
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
