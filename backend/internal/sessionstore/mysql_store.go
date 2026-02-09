package sessionstore

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// MySQLStore implements fiber.Storage backed by MySQL.
type MySQLStore struct {
	db *sqlx.DB
}

func NewMySQLStore(db *sqlx.DB) *MySQLStore {
	return &MySQLStore{db: db}
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
			log.Printf("failed to delete expired session %s: %v", key, err)
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
	return nil
}
