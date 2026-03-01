package models

import (
	"errors"
	"time"
)

var ErrDuplicate = errors.New("duplicate entry")

type Person struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description"`
	CreatedBy   int64     `db:"created_by_user_id" json:"created_by_user_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
