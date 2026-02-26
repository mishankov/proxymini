package requestlog

import (
	"github.com/jmoiron/sqlx"
	"github.com/platforma-dev/platforma/log"
)

type RequestLogService struct {
	db           *sqlx.DB
	requestLogCh chan RequestLog
}

func NewRequestLogService(db *sqlx.DB) *RequestLogService {
	ch := make(chan RequestLog)
	rls := &RequestLogService{db: db, requestLogCh: ch}
	go func() {
		for l := range ch {
			err := rls.save(l)
			if err != nil {
				log.Error("failed to save request log from channel", "error", err)
			}
		}
	}()
	return rls
}

func (rls *RequestLogService) GetList() ([]RequestLog, error) {
	var res []RequestLog

	rows, err := rls.db.Queryx("SELECT * FROM request_log ORDER BY time DESC")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var rl RequestLog

		err = rows.StructScan(&rl)
		if err != nil {
			return nil, err
		}

		res = append(res, rl)
	}

	return res, nil
}

func (rls *RequestLogService) save(rl RequestLog) error {
	_, err := rls.db.Exec(
		"INSERT INTO request_log (id, time, elapsed_ms, method, proxy_url, url, request_headers, request_body, status, response_headers, response_body) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
		rl.ID, rl.Time, rl.ElapsedMS, rl.Method, rl.ProxyURL, rl.URL, rl.RequestHeaders, rl.RequestBody, rl.Status, rl.ResponseHeaders, rl.ResponseBody,
	)

	return err
}

func (rls *RequestLogService) Save(rl RequestLog) error {
	rls.requestLogCh <- rl
	return nil
}

func (rls *RequestLogService) DeleteAll() error {
	_, err := rls.db.Exec("DELETE FROM request_log")

	return err
}

func (rls *RequestLogService) DeleteOlderThan(cutoffTime int64) error {
	_, err := rls.db.Exec("DELETE FROM request_log WHERE time < ?", cutoffTime)
	return err
}
