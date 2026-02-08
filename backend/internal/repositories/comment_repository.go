package repositories

import "familyjournal/backend/internal/models"

func (r *Repository) CreateComment(comment *models.Comment) error {
	res, err := r.DB.Exec(`INSERT INTO comments (post_id, user_id, text, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`, comment.PostID, comment.UserID, comment.Text)
	if err != nil {
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
	return r.DB.Get(comment, `SELECT c.id, c.post_id, c.user_id, c.text, c.created_at, c.updated_at, u.email AS author_email
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = ?`, id)
}

func (r *Repository) UpdateComment(comment *models.Comment) error {
	_, err := r.DB.Exec(`UPDATE comments SET text = ?, updated_at = NOW() WHERE id = ? AND user_id = ?`, comment.Text, comment.ID, comment.UserID)
	return err
}

func (r *Repository) DeleteComment(id, userID int64) error {
	_, err := r.DB.Exec(`DELETE FROM comments WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
