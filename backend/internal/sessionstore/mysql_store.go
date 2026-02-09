package sessionstore

import (
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

const cleanupInterval = time.Hour

type sessionDB interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// MySQLStore implements fiber.Storage backed by MySQL.
type MySQLStore struct {
	db           sessionDB
	cleanupDone  chan struct{}
	cleanupStop  chan struct{}
	cleanupClose sync.Once
}

func NewMySQLStore(db *sqlx.DB) *MySQLStore {
	store := &MySQLStore{
		db:          db,
		cleanupDone: make(chan struct{}),
		cleanupStop: make(chan struct{}),
	}
	go store.cleanupExpiredSessions()
	return store
}

func (s *MySQLStore) cleanupExpiredSessions() {
	ticker := time.NewTicker(cleanupInterval)
	defer func() {
		ticker.Stop()
		close(s.cleanupDone)
	}()

	for {
		select {
		case <-ticker.C:
			if _, err := s.db.Exec("DELETE FROM session_store WHERE expires_at IS NOT NULL AND expires_at < ?", time.Now().UTC()); err != nil {
				log.Printf("failed to cleanup expired sessions: %v", err)
			}
		case <-s.cleanupStop:
			return
		}
	}
}

func (s *MySQLStore) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, nil
	}

	var row struct {
		Data      []byte     `db:"data"`
		ExpiresAt *time.Time `db:"expires_at"`
	}
	err := s.db.Get(&row, "SELECT data, expires_at FROM session_store WHERE id = ?", key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if row.ExpiresAt != nil && row.ExpiresAt.Before(time.Now().UTC()) {
		if _, err := s.db.Exec("DELETE FROM session_store WHERE id = ?", key); err != nil {
			log.Printf("failed to delete expired session: %v", err)
		}
		return nil, nil
	}

	return row.Data, nil
}

func (s *MySQLStore) Set(key string, val []byte, exp time.Duration) error {
	if key == "" || len(val) == 0 {
		return nil
	}

	var expiresAt *time.Time
	if exp > 0 {
		expiry := time.Now().UTC().Add(exp)
		expiresAt = &expiry
	}

	_, err := s.db.Exec(`
		INSERT INTO session_store (id, data, expires_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE data = VALUES(data), expires_at = VALUES(expires_at)
	`, key, val, expiresAt)
	return err
}

func (s *MySQLStore) Delete(key string) error {
	if key == "" {
		return nil
	}
	_, err := s.db.Exec("DELETE FROM session_store WHERE id = ?", key)
	return err
}

func (s *MySQLStore) Reset() error {
	_, err := s.db.Exec("TRUNCATE TABLE session_store")
	return err
}

func (s *MySQLStore) Close() error {
	s.cleanupClose.Do(func() {
		close(s.cleanupStop)
		<-s.cleanupDone
	})
	return nil
}
