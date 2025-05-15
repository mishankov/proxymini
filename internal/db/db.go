package db

import (
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
    method VARCHAR(100),    
    url TEXT,               
    request_headers TEXT,   
    request_body TEXT,      
    status INT NOT NULL,    
    response_headers TEXT,  
    response_body TEXT      
);`)

	return err
}
