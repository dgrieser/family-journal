package db

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
}

func New(dsn string, cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(cfg.MaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	const (
		maxPingAttempts = 30
		pingInterval    = 2 * time.Second
	)
	var pingErr error
	for attempt := 1; attempt <= maxPingAttempts; attempt++ {
		pingErr = db.Ping()
		if pingErr == nil {
			return db, nil
		}
		time.Sleep(pingInterval)
	}
	_ = db.Close()
	return nil, pingErr
}
