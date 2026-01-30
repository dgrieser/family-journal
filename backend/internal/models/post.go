package models

import (
	"time"
)

type Post struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	UserID      uint         `gorm:"not null" json:"user_id"`
	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Date        time.Time    `gorm:"type:date;not null;index" json:"date"`
	Text        string       `gorm:"type:text;not null" json:"text"`
	Hashtags    []Hashtag    `gorm:"many2many:post_hashtags;" json:"hashtags"`
	Mentions    []Person     `gorm:"many2many:mentions;" json:"mentions"`
	Attachments []Attachment `gorm:"foreignKey:PostID" json:"attachments"`
	Comments    []Comment    `gorm:"foreignKey:PostID" json:"comments"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Hashtag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Attachment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PostID      uint      `gorm:"not null" json:"post_id"`
	FileName    string    `gorm:"not null" json:"file_name"`
	FileType    string    `gorm:"not null" json:"file_type"`
	FileSize    int64     `gorm:"not null" json:"file_size"`
	StoragePath string    `gorm:"not null" json:"storage_path"`
	CreatedAt   time.Time `json:"created_at"`
}
