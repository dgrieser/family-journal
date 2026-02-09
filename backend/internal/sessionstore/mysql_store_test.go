package sessionstore

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"
)

type fakeResult struct{}

func (f fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (f fakeResult) RowsAffected() (int64, error) { return 1, nil }

type fakeDB struct {
	getFn       func(dest interface{}, query string, args ...interface{}) error
	execFn      func(query string, args ...interface{}) (sql.Result, error)
	execQueries []string
	execArgs    [][]interface{}
}

func (f *fakeDB) Get(dest interface{}, query string, args ...interface{}) error {
	if f.getFn != nil {
		return f.getFn(dest, query, args...)
	}
	return nil
}

func (f *fakeDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	f.execQueries = append(f.execQueries, query)
	f.execArgs = append(f.execArgs, args)
	if f.execFn != nil {
		return f.execFn(query, args...)
	}
	return fakeResult{}, nil
}

func TestGetNotFound(t *testing.T) {
	db := &fakeDB{getFn: func(dest interface{}, query string, args ...interface{}) error {
		return sql.ErrNoRows
	}}
	store := &MySQLStore{db: db}

	got, err := store.Get("missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil data, got %v", got)
	}
}

func TestGetExpiredSessionDeletesRow(t *testing.T) {
	expiredAt := time.Now().UTC().Add(-time.Minute)
	db := &fakeDB{getFn: func(dest interface{}, query string, args ...interface{}) error {
		row := dest.(*struct {
			Data      []byte     `db:"data"`
			ExpiresAt *time.Time `db:"expires_at"`
		})
		row.Data = []byte("payload")
		row.ExpiresAt = &expiredAt
		return nil
	}}
	store := &MySQLStore{db: db}

	got, err := store.Get("session-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil data for expired session, got %v", got)
	}
	if len(db.execQueries) != 1 || db.execQueries[0] != "DELETE FROM session_store WHERE id = ?" {
		t.Fatalf("unexpected cleanup query calls: %v", db.execQueries)
	}
	if len(db.execArgs[0]) != 1 || db.execArgs[0][0] != "session-key" {
		t.Fatalf("unexpected cleanup query args: %v", db.execArgs[0])
	}
}

func TestSetUsesUpsert(t *testing.T) {
	db := &fakeDB{}
	store := &MySQLStore{db: db}

	err := store.Set("k", []byte("v"), 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(db.execQueries) != 1 {
		t.Fatalf("expected one exec call, got %d", len(db.execQueries))
	}
	if db.execQueries[0] == "" {
		t.Fatal("expected non-empty query")
	}
	if len(db.execArgs[0]) != 3 {
		t.Fatalf("expected 3 args, got %d", len(db.execArgs[0]))
	}
	if db.execArgs[0][0] != "k" {
		t.Fatalf("unexpected key arg: %v", db.execArgs[0][0])
	}
	if !reflect.DeepEqual(db.execArgs[0][1], []byte("v")) {
		t.Fatalf("unexpected value arg: %v", db.execArgs[0][1])
	}
	if _, ok := db.execArgs[0][2].(*time.Time); !ok {
		t.Fatalf("expected expiresAt *time.Time, got %T", db.execArgs[0][2])
	}
}

func TestDeleteExecutesDelete(t *testing.T) {
	db := &fakeDB{}
	store := &MySQLStore{db: db}

	if err := store.Delete("abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(db.execQueries) != 1 || db.execQueries[0] != "DELETE FROM session_store WHERE id = ?" {
		t.Fatalf("unexpected delete query calls: %v", db.execQueries)
	}
}

func TestResetUsesTruncate(t *testing.T) {
	db := &fakeDB{}
	store := &MySQLStore{db: db}

	if err := store.Reset(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(db.execQueries) != 1 || db.execQueries[0] != "TRUNCATE TABLE session_store" {
		t.Fatalf("unexpected reset query calls: %v", db.execQueries)
	}
}

func TestGetReturnsDBError(t *testing.T) {
	dbErr := errors.New("db down")
	db := &fakeDB{getFn: func(dest interface{}, query string, args ...interface{}) error {
		return dbErr
	}}
	store := &MySQLStore{db: db}

	_, err := store.Get("k")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}
