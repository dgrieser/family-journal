package models

import "time"

type Post struct {
	ID          int64        `db:"id" json:"id"`
	UserID      int64        `db:"user_id" json:"user_id"`
	Date        time.Time    `db:"date" json:"date"`
	Text        string       `db:"text" json:"text"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
	Hashtags    []Hashtag    `json:"hashtags"`
	Persons     []Person     `json:"persons"`
	Comments    []Comment    `json:"comments"`
	Attachments []Attachment `json:"attachments"`
}
