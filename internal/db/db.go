package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func Connect(name string) (*sqlx.DB, error) {
	return sqlx.Connect("sqlite", name)
}

func Init(db *sqlx.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS request_log (
    id TEXT PRIMARY KEY,  
    time BIGINT NOT NULL,   
    elapsed_ms BIGINT NOT NULL DEFAULT 0,
    method TEXT,  
    proxy_url TEXT,
    url TEXT,               
    request_headers TEXT,   
    request_body TEXT,      
    status INT NOT NULL,    
    response_headers TEXT,  
    response_body TEXT      
);`)
	if err != nil {
		return fmt.Errorf("create request_log table: %w", err)
	}

	// Create index on time column for efficient log retention cleanup
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_request_log_time ON request_log(time)"); err != nil {
		return fmt.Errorf("create index on request_log.time: %w", err)
	}

	return nil
}
