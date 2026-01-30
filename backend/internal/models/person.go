package models

import (
	"time"
)

type Person struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"uniqueIndex;not null" json:"name"`
	Description       string    `json:"description"`
	CreatedByUserID   uint      `json:"created_by_user_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (Person) TableName() string {
	return "persons"
}
