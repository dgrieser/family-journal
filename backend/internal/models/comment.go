package models

import "time"

type Comment struct {
	ID          int64       `db:"id" json:"id"`
	PostID      int64       `db:"post_id" json:"post_id"`
	UserID      int64       `db:"user_id" json:"-"`
	Text        string      `db:"text" json:"text"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
	AuthorEmail string      `db:"author_email" json:"-"`
	User        CommentUser `json:"user"`
}

// HydrateUser sets the nested user payload from legacy flat author fields.
func (c *Comment) HydrateUser() {
	c.User = CommentUser{
		ID:    c.UserID,
		Email: c.AuthorEmail,
	}
}

// CommentUser contains the minimum author fields needed by comment responses.
type CommentUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}
