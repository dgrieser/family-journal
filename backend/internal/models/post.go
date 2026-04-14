package models

import "time"

type PostUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type Post struct {
	ID          int64        `db:"id" json:"id"`
	UserID      int64        `db:"user_id" json:"user_id"`
	Date        time.Time    `db:"date" json:"date"`
	Text        string       `db:"text" json:"text"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
	AuthorEmail string       `db:"author_email" json:"-"`
	User        *PostUser    `json:"user"`
	Hashtags    []Hashtag    `json:"hashtags"`
	Persons     []Person     `json:"persons"`
	Comments    []Comment    `json:"comments"`
	Attachments []Attachment `json:"attachments"`
}

func (p *Post) HydrateUser() {
	p.User = &PostUser{ID: p.UserID, Email: p.AuthorEmail}
}
