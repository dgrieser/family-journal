package repositories

import "familyjournal/backend/internal/models"

func (r *Repository) CreateAttachment(att *models.Attachment) error {
	res, err := r.DB.Exec(`INSERT INTO attachments (post_id, file_name, file_type, file_size, created_at)
		VALUES (?, ?, ?, ?, NOW())`, att.PostID, att.FileName, att.FileType, att.FileSize)
	if err != nil {
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
	return r.DB.Get(att, `SELECT id, post_id, file_name, file_type, file_size, created_at FROM attachments WHERE id = ?`, id)
}

func (r *Repository) GetAttachmentByID(id int64, ownerFilter *int64) (*models.Attachment, error) {
	var attachment models.Attachment
	query := `SELECT a.id, a.post_id, a.file_name, a.file_type, a.file_size, a.created_at
		FROM attachments a
		JOIN posts p ON p.id = a.post_id
		WHERE a.id = ?`
	args := []interface{}{id}
	if ownerFilter != nil {
		query += ` AND p.user_id = ?`
		args = append(args, *ownerFilter)
	}
	if err := r.DB.Get(&attachment, query, args...); err != nil {
		return nil, err
	}
	return &attachment, nil
}
