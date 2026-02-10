package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"familyjournal/backend/internal/models"

	"github.com/jmoiron/sqlx"
)

var errPostNotFoundOrForbidden = fmt.Errorf("post not found or access denied")

func (r *Repository) SavePostWithRelations(ownerID int64, ownerFilter *int64, post *models.Post, tagNames, personNames []string) (err error) {
	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			err = commitErr
		}
	}()

	if post.ID == 0 {
		post.UserID = ownerID
		query := `INSERT INTO posts (user_id, date, text, category, mood, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, NOW(), NOW())`
		res, execErr := tx.Exec(query, post.UserID, post.Date, post.Text, post.Category, post.Mood)
		if execErr != nil {
			return execErr
		}
		id, err := lastInsertID(res)
		if err != nil {
			return err
		}
		post.ID = id
	} else {
		post.UserID = ownerID
		query := `UPDATE posts SET text = ?, category = ?, mood = ?, updated_at = NOW() WHERE id = ?`
		args := []interface{}{post.Text, post.Category, post.Mood, post.ID}
		if ownerFilter != nil {
			query += ` AND user_id = ?`
			args = append(args, *ownerFilter)
		}
		res, execErr := tx.Exec(query, args...)
		if execErr != nil {
			return execErr
		}
		rowsAffected, rowsErr := res.RowsAffected()
		if rowsErr != nil {
			return rowsErr
		}
		if rowsAffected == 0 {
			return errPostNotFoundOrForbidden
		}
	}

	var tagModels []models.Hashtag
	for _, tag := range tagNames {
		model, execErr := findOrCreateHashtagTx(tx, tag)
		if execErr != nil {
			return execErr
		}
		tagModels = append(tagModels, *model)
	}

	var personModels []models.Person
	for _, name := range personNames {
		model, execErr := findOrCreatePersonTx(tx, ownerID, name)
		if execErr != nil {
			return execErr
		}
		personModels = append(personModels, *model)
	}

	if execErr := replacePostTagsTx(tx, post.ID, tagModels); execErr != nil {
		return execErr
	}
	if execErr := replacePostMentionsTx(tx, post.ID, personModels); execErr != nil {
		return execErr
	}
	return nil
}

func (r *Repository) CreatePost(post *models.Post) error {
	query := `INSERT INTO posts (user_id, date, text, category, mood, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())`
	res, err := r.DB.Exec(query, post.UserID, post.Date, post.Text, post.Category, post.Mood)
	if err != nil {
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
	post.ID = id
	return nil
}

func (r *Repository) UpdatePost(post *models.Post) error {
	_, err := r.DB.Exec(`UPDATE posts SET text = ?, category = ?, mood = ?, updated_at = NOW() WHERE id = ? AND user_id = ?`,
		post.Text, post.Category, post.Mood, post.ID, post.UserID)
	return err
}

func (r *Repository) DeletePost(id int64, ownerFilter *int64) error {
	query := `DELETE FROM posts WHERE id = ?`
	args := []interface{}{id}
	if ownerFilter != nil {
		query += ` AND user_id = ?`
		args = append(args, *ownerFilter)
	}
	_, err := r.DB.Exec(query, args...)
	return err
}

func (r *Repository) GetPost(id int64, ownerFilter *int64) (*models.Post, error) {
	var post models.Post
	query := `SELECT id, user_id, date, text, category, mood, created_at, updated_at FROM posts WHERE id = ?`
	args := []interface{}{id}
	if ownerFilter != nil {
		query += ` AND user_id = ?`
		args = append(args, *ownerFilter)
	}
	if err := r.DB.Get(&post, query, args...); err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *Repository) ListPosts(ownerFilter *int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error) {
	base := `SELECT DISTINCT p.id, p.user_id, p.date, p.text, p.category, p.mood, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN post_hashtags ph ON ph.post_id = p.id
		LEFT JOIN hashtags h ON h.id = ph.hashtag_id
		LEFT JOIN mentions m ON m.post_id = p.id
		LEFT JOIN persons pe ON pe.id = m.person_id
		WHERE p.date = ?`
	args := []interface{}{date}
	if ownerFilter != nil {
		base += ` AND p.user_id = ?`
		args = append(args, *ownerFilter)
	}
	if len(hashtags) > 0 {
		placeholders := strings.Repeat("?,", len(hashtags))
		placeholders = strings.TrimSuffix(placeholders, ",")
		base += fmt.Sprintf(" AND h.name IN (%s)", placeholders)
		for _, tag := range hashtags {
			args = append(args, strings.ToLower(tag))
		}
	}
	if len(persons) > 0 {
		placeholders := strings.Repeat("?,", len(persons))
		placeholders = strings.TrimSuffix(placeholders, ",")
		base += fmt.Sprintf(" AND pe.name IN (%s)", placeholders)
		for _, person := range persons {
			args = append(args, person)
		}
	}
	if search != "" {
		base += " AND p.text LIKE ?"
		args = append(args, "%"+search+"%")
	}
	base += " ORDER BY p.created_at DESC"
	var posts []models.Post
	if err := r.DB.Select(&posts, base, args...); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *Repository) ReplacePostTags(postID int64, tags []models.Hashtag) error {
	if _, err := r.DB.Exec(`DELETE FROM post_hashtags WHERE post_id = ?`, postID); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := r.DB.Exec(`INSERT INTO post_hashtags (post_id, hashtag_id) VALUES (?, ?)`, postID, tag.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ReplacePostMentions(postID int64, persons []models.Person) error {
	if _, err := r.DB.Exec(`DELETE FROM mentions WHERE post_id = ?`, postID); err != nil {
		return err
	}
	for _, person := range persons {
		if _, err := r.DB.Exec(`INSERT INTO mentions (post_id, person_id) VALUES (?, ?)`, postID, person.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListPostComments(postID int64) ([]models.Comment, error) {
	var comments []models.Comment
	query := `SELECT c.id, c.post_id, c.user_id, c.text, c.created_at, c.updated_at, u.email AS author_email
		FROM comments c JOIN users u ON u.id = c.user_id WHERE c.post_id = ? ORDER BY c.created_at ASC`
	if err := r.DB.Select(&comments, query, postID); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *Repository) ListPostTags(postID int64) ([]models.Hashtag, error) {
	var tags []models.Hashtag
	query := `SELECT h.id, h.name, h.created_at FROM hashtags h
		JOIN post_hashtags ph ON ph.hashtag_id = h.id WHERE ph.post_id = ?`
	if err := r.DB.Select(&tags, query, postID); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *Repository) ListPostPersons(postID int64) ([]models.Person, error) {
	var persons []models.Person
	query := `SELECT p.id, p.name, p.description, p.created_by_user_id, p.created_at, p.updated_at FROM persons p
		JOIN mentions m ON m.person_id = p.id WHERE m.post_id = ?`
	if err := r.DB.Select(&persons, query, postID); err != nil {
		return nil, err
	}
	return persons, nil
}

func (r *Repository) ListPostAttachments(postID int64) ([]models.Attachment, error) {
	var attachments []models.Attachment
	query := `SELECT id, post_id, file_name, file_type, file_size, url, created_at FROM attachments WHERE post_id = ?`
	if err := r.DB.Select(&attachments, query, postID); err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *Repository) ListPersonsForPosts(postIDs []int64) (map[int64][]models.Person, error) {
	results := make(map[int64][]models.Person)
	if len(postIDs) == 0 {
		return results, nil
	}
	query, args, err := sqlx.In(`SELECT m.post_id, p.id, p.name, p.description, p.created_by_user_id, p.created_at, p.updated_at
		FROM mentions m
		JOIN persons p ON p.id = m.person_id
		WHERE m.post_id IN (?)`, postIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB.Rebind(query)
	type row struct {
		PostID      int64     `db:"post_id"`
		ID          int64     `db:"id"`
		Name        string    `db:"name"`
		Description *string   `db:"description"`
		CreatedBy   int64     `db:"created_by_user_id"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}
	var rows []row
	if err := r.DB.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	for _, item := range rows {
		results[item.PostID] = append(results[item.PostID], models.Person{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			CreatedBy:   item.CreatedBy,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}
	return results, nil
}

func (r *Repository) ListCommentsForPosts(postIDs []int64) (map[int64][]models.Comment, error) {
	results := make(map[int64][]models.Comment)
	if len(postIDs) == 0 {
		return results, nil
	}
	query, args, err := sqlx.In(`SELECT c.post_id, c.id, c.user_id, c.text, c.created_at, c.updated_at, u.email AS author_email
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.post_id IN (?)
		ORDER BY c.created_at ASC`, postIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB.Rebind(query)
	type row struct {
		PostID      int64     `db:"post_id"`
		ID          int64     `db:"id"`
		UserID      int64     `db:"user_id"`
		Text        string    `db:"text"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
		AuthorEmail string    `db:"author_email"`
	}
	var rows []row
	if err := r.DB.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	for _, item := range rows {
		results[item.PostID] = append(results[item.PostID], models.Comment{
			ID:          item.ID,
			PostID:      item.PostID,
			UserID:      item.UserID,
			Text:        item.Text,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			AuthorEmail: item.AuthorEmail,
		})
	}
	return results, nil
}

func (r *Repository) ListAttachmentsForPosts(postIDs []int64) (map[int64][]models.Attachment, error) {
	results := make(map[int64][]models.Attachment)
	if len(postIDs) == 0 {
		return results, nil
	}
	query, args, err := sqlx.In(`SELECT post_id, id, file_name, file_type, file_size, url, created_at
		FROM attachments WHERE post_id IN (?)`, postIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB.Rebind(query)
	type row struct {
		PostID    int64     `db:"post_id"`
		ID        int64     `db:"id"`
		FileName  string    `db:"file_name"`
		FileType  string    `db:"file_type"`
		FileSize  int64     `db:"file_size"`
		URL       string    `db:"url"`
		CreatedAt time.Time `db:"created_at"`
	}
	var rows []row
	if err := r.DB.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	for _, item := range rows {
		results[item.PostID] = append(results[item.PostID], models.Attachment{
			ID:        item.ID,
			PostID:    item.PostID,
			FileName:  item.FileName,
			FileType:  item.FileType,
			FileSize:  item.FileSize,
			URL:       item.URL,
			CreatedAt: item.CreatedAt,
		})
	}
	return results, nil
}

func findOrCreatePersonTx(tx *sqlx.Tx, userID int64, name string) (*models.Person, error) {
	var person models.Person
	if err := tx.Get(&person, `SELECT id, name, description, created_by_user_id, created_at, updated_at
		FROM persons WHERE created_by_user_id = ? AND name = ?`, userID, name); err == nil {
		return &person, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	person = models.Person{Name: name, CreatedBy: userID}
	res, err := tx.Exec(`INSERT INTO persons (name, description, created_by_user_id, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`, person.Name, person.Description, person.CreatedBy)
	if err != nil {
		return nil, err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return nil, err
	}
	if err := tx.Get(&person, `SELECT id, name, description, created_by_user_id, created_at, updated_at
		FROM persons WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &person, nil
}

func replacePostTagsTx(tx *sqlx.Tx, postID int64, tags []models.Hashtag) error {
	if _, err := tx.Exec(`DELETE FROM post_hashtags WHERE post_id = ?`, postID); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := tx.Exec(`INSERT INTO post_hashtags (post_id, hashtag_id) VALUES (?, ?)`, postID, tag.ID); err != nil {
			return err
		}
	}
	return nil
}

func replacePostMentionsTx(tx *sqlx.Tx, postID int64, persons []models.Person) error {
	if _, err := tx.Exec(`DELETE FROM mentions WHERE post_id = ?`, postID); err != nil {
		return err
	}
	for _, person := range persons {
		if _, err := tx.Exec(`INSERT INTO mentions (post_id, person_id) VALUES (?, ?)`, postID, person.ID); err != nil {
			return err
		}
	}
	return nil
}
