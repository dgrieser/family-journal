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
	if err := r.loadCommentWithAuthor(comment, id, nil); err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpdateComment(comment *models.Comment, ownerFilter *int64) error {
	query := `UPDATE comments SET text = ?, updated_at = NOW() WHERE id = ?`
	args := []interface{}{comment.Text, comment.ID}
	if ownerFilter != nil {
		query += ` AND user_id = ?`
		args = append(args, *ownerFilter)
	}
	if _, err := r.DB.Exec(query, args...); err != nil {
		return err
	}
	if err := r.loadCommentWithAuthor(comment, comment.ID, ownerFilter); err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteComment(id int64, ownerFilter *int64) error {
	query := `DELETE FROM comments WHERE id = ?`
	args := []interface{}{id}
	if ownerFilter != nil {
		query += ` AND user_id = ?`
		args = append(args, *ownerFilter)
	}
	_, err := r.DB.Exec(query, args...)
	return err
}

func (r *Repository) loadCommentWithAuthor(comment *models.Comment, commentID int64, ownerFilter *int64) error {
	query := `SELECT c.id, c.post_id, c.user_id, c.text, c.created_at, c.updated_at, u.email AS author_email
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = ?`
	args := []interface{}{commentID}
	if ownerFilter != nil {
		query += ` AND c.user_id = ?`
		args = append(args, *ownerFilter)
	}
	if err := r.DB.Get(comment, query, args...); err != nil {
		return err
	}
	comment.HydrateUser()
	return nil
}
