package requestlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/platforma-dev/platforma/log"
)

type RequestLogHandler struct {
	rlSvc *RequestLogService
}

func NewRequestLogHandler(rlSvc *RequestLogService) *RequestLogHandler {
	return &RequestLogHandler{rlSvc: rlSvc}
}

func (rlh *RequestLogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		res, err := rlh.rlSvc.GetList()
		if err != nil {
			handleError(w, fmt.Errorf("app error: %w", err), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(res)
		if err != nil {
			handleError(w, fmt.Errorf("app error: %w", err), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_, err = w.Write(data)
		if err != nil {
			handleError(w, fmt.Errorf("app error: %w", err), http.StatusInternalServerError)
			return
		}

	case "DELETE":
		err := rlh.rlSvc.DeleteAll()
		if err != nil {
			handleError(w, fmt.Errorf("deleting request logs: %w", err), http.StatusInternalServerError)
		}
	}
}

func handleError(w http.ResponseWriter, err error, status int) {
	log.ErrorContext(context.Background(), "request handling error", "status", status, "error", err)
	// w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}
