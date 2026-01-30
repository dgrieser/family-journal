package models

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password_hash" json:"-"`
	Role      string    `db:"role" json:"role"`
	Active    bool      `db:"active" json:"active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Person struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description"`
	CreatedBy   int64     `db:"created_by_user_id" json:"created_by_user_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type Post struct {
	ID          int64        `db:"id" json:"id"`
	UserID      int64        `db:"user_id" json:"user_id"`
	Date        time.Time    `db:"date" json:"date"`
	Text        string       `db:"text" json:"text"`
	Category    *string      `db:"category" json:"category"`
	Mood        *string      `db:"mood" json:"mood"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
	Hashtags    []Hashtag    `json:"hashtags"`
	Persons     []Person     `json:"persons"`
	Comments    []Comment    `json:"comments"`
	Attachments []Attachment `json:"attachments"`
}

type Comment struct {
	ID          int64     `db:"id" json:"id"`
	PostID      int64     `db:"post_id" json:"post_id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	Text        string    `db:"text" json:"text"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	AuthorEmail string    `db:"author_email" json:"author_email"`
}

type Hashtag struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Attachment struct {
	ID        int64     `db:"id" json:"id"`
	PostID    int64     `db:"post_id" json:"post_id"`
	FileName  string    `db:"file_name" json:"file_name"`
	FileType  string    `db:"file_type" json:"file_type"`
	FileSize  int64     `db:"file_size" json:"file_size"`
	URL       string    `db:"url" json:"url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
