package repositories

import (
	"database/sql"

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
