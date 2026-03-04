package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
}

func lastInsertID(result sql.Result) (int64, error) {
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func New(db *sqlx.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) beginReadSnapshotTx() (*sqlx.Tx, error) {
	return r.DB.BeginTxx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func resolveDuplicateInsert(insertErr error, fetchExisting func() error) error {
	if !isDuplicateKeyError(insertErr) {
		return insertErr
	}
	if err := fetchExisting(); err != nil {
		return err
	}
	return nil
}
