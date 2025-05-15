package requestlog

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RequestLog struct {
	ID              string `json:"id"`
	Time            int64  `json:"time"`
	Method          string `json:"method"`
	URL             string `json:"url"`
	RequestHeaders  string `db:"request_headers" json:"requestHeaders"`
	RequestBody     string `db:"request_body" json:"requestBody"`
	Status          int    `json:"status"`
	ResponseHeaders string `db:"response_headers" json:"responseHeaders"`
	ResponseBody    string `db:"response_body" json:"responseBody"`
}

func New(method, URL string, requestHeaders http.Header, requestBody string, status int, responseHeaders http.Header, responseBody string) RequestLog {
	requestHeadersBytes, _ := json.Marshal(requestHeaders)
	responseHeadersBytes, _ := json.Marshal(responseHeaders)

	return RequestLog{
		ID:              uuid.NewString(),
		Time:            time.Now().UTC().Unix(),
		Method:          method,
		URL:             URL,
		RequestHeaders:  string(requestHeadersBytes),
		RequestBody:     requestBody,
		Status:          status,
		ResponseHeaders: string(responseHeadersBytes),
		ResponseBody:    responseBody,
	}
}
