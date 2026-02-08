package repositories

import "familyjournal/backend/internal/models"

func (r *Repository) CreateAttachment(att *models.Attachment) error {
	res, err := r.DB.Exec(`INSERT INTO attachments (post_id, file_name, file_type, file_size, url, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())`, att.PostID, att.FileName, att.FileType, att.FileSize, att.URL)
	if err != nil {
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
	return r.DB.Get(att, `SELECT id, post_id, file_name, file_type, file_size, url, created_at FROM attachments WHERE id = ?`, id)
}

func (r *Repository) GetAttachmentByName(userID int64, name string) (*models.Attachment, error) {
	var attachment models.Attachment
	query := `SELECT a.id, a.post_id, a.file_name, a.file_type, a.file_size, a.url, a.created_at
		FROM attachments a
		JOIN posts p ON p.id = a.post_id
		WHERE a.file_name = ? AND p.user_id = ?`
	if err := r.DB.Get(&attachment, query, name, userID); err != nil {
		return nil, err
	}
	return &attachment, nil
}
