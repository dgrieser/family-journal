package models

import "time"

type Attachment struct {
	ID          int64     `db:"id" json:"id"`
	PostID      int64     `db:"post_id" json:"post_id"`
	FileName    string    `db:"file_name" json:"file_name"`
	FileType    string    `db:"file_type" json:"file_type"`
	FileSize    int64     `db:"file_size" json:"file_size"`
	StoragePath string    `db:"storage_path" json:"storage_path"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
