package services

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/mishankov/proxymini/internal/requestlog"
)

type RequestLogService struct {
	db           *sqlx.DB
	requestLogCh chan requestlog.RequestLog
}

func NewRequestLogService(db *sqlx.DB) *RequestLogService {
	ch := make(chan requestlog.RequestLog)
	rls := &RequestLogService{db: db, requestLogCh: ch}
	go func() {
		for l := range ch {
			err := rls.save(l)
			if err != nil {
				log.Println("Error saving request log from channel:", err)
			}
		}
	}()
	return rls
}

func (rls *RequestLogService) GetList() ([]requestlog.RequestLog, error) {
	var res []requestlog.RequestLog

	rows, err := rls.db.Queryx("SELECT * FROM request_log ORDER BY time DESC")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var rl requestlog.RequestLog

		err = rows.StructScan(&rl)
		if err != nil {
			return nil, err
		}

		res = append(res, rl)
	}

	return res, nil
}

func (rls *RequestLogService) save(rl requestlog.RequestLog) error {
	_, err := rls.db.Exec(
		"INSERT INTO request_log (id, time, method, url, request_headers, request_body, status, response_headers, response_body) VALUES (?,?,?,?,?,?,?,?,?)",
		rl.ID, rl.Time, rl.Method, rl.URL, rl.RequestHeaders, rl.RequestBody, rl.Status, rl.ResponseHeaders, rl.ResponseBody,
	)

	return err
}

func (rls *RequestLogService) Save(rl requestlog.RequestLog) error {
	rls.requestLogCh <- rl
	return nil
}

func (rls *RequestLogService) DeleteAll() error {
	_, err := rls.db.Exec("DELETE FROM request_log")

	return err
}
