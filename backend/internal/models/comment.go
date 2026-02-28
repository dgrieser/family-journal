package models

import "time"

type Comment struct {
	ID          int64       `db:"id" json:"id"`
	PostID      int64       `db:"post_id" json:"post_id"`
	UserID      int64       `db:"user_id" json:"user_id"`
	Text        string      `db:"text" json:"text"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
	AuthorEmail string      `db:"author_email" json:"author_email"`
	User        CommentUser `json:"user"`
}

type CommentUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}
