package models

import (
	"errors"
	"time"
)

var ErrDuplicate = errors.New("duplicate entry")
var ErrForbidden = errors.New("forbidden")

type PersonCreator struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type Person struct {
	ID           int64          `db:"id" json:"id"`
	Name         string         `db:"name" json:"name"`
	Description  *string        `db:"description" json:"description"`
	CreatedBy    int64          `db:"created_by_user_id" json:"created_by_user_id"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
	CreatorEmail string         `db:"creator_email" json:"-"`
	Creator      *PersonCreator `json:"creator"`
}

func (p *Person) HydrateCreator() {
	p.Creator = &PersonCreator{ID: p.CreatedBy, Email: p.CreatorEmail}
}
