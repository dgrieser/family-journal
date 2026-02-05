package models

import (
	"time"
)

type Comment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;type:int" json:"id"`
	PostID    uint      `gorm:"type:int;not null" json:"post_id"`
	UserID    uint      `gorm:"type:int;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Text      string    `gorm:"type:text;not null" json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
