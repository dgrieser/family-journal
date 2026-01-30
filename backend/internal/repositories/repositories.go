package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"familyjournal/backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (email, password_hash, role, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())`
	res, err := r.DB.Exec(query, user.Email, user.Password, user.Role, user.Active)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	user.ID = id
	return nil
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, role, active, created_at, updated_at FROM users WHERE email = ?`
	if err := r.DB.Get(&user, query, email); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, role, active, created_at, updated_at FROM users WHERE id = ?`
	if err := r.DB.Get(&user, query, id); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUserProfile(id int64, email string) error {
	_, err := r.DB.Exec(`UPDATE users SET email = ?, updated_at = NOW() WHERE id = ?`, email, id)
	return err
}

func (r *Repository) ListUsers() ([]models.User, error) {
	var users []models.User
	query := `SELECT id, email, role, active, created_at, updated_at FROM users ORDER BY created_at DESC`
	if err := r.DB.Select(&users, query); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) UpdateUserRole(id int64, role string) error {
	_, err := r.DB.Exec(`UPDATE users SET role = ?, updated_at = NOW() WHERE id = ?`, role, id)
	return err
}

func (r *Repository) UpdateUserActive(id int64, active bool) error {
	_, err := r.DB.Exec(`UPDATE users SET active = ?, updated_at = NOW() WHERE id = ?`, active, id)
	return err
}

func (r *Repository) CreatePerson(person *models.Person) error {
	query := `INSERT INTO persons (name, description, created_by_user_id, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`
	res, err := r.DB.Exec(query, person.Name, person.Description, person.CreatedBy)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	person.ID = id
	return nil
}

func (r *Repository) UpdatePerson(person *models.Person) error {
	_, err := r.DB.Exec(`UPDATE persons SET name = ?, description = ?, updated_at = NOW() WHERE id = ? AND created_by_user_id = ?`,
		person.Name, person.Description, person.ID, person.CreatedBy)
	return err
}

func (r *Repository) DeletePerson(id, userID int64) error {
	_, err := r.DB.Exec(`DELETE FROM persons WHERE id = ? AND created_by_user_id = ?`, id, userID)
	return err
}

func (r *Repository) ListPersons(userID int64) ([]models.Person, error) {
	var persons []models.Person
	query := `SELECT id, name, description, created_by_user_id, created_at, updated_at FROM persons WHERE created_by_user_id = ? ORDER BY name ASC`
	if err := r.DB.Select(&persons, query, userID); err != nil {
		return nil, err
	}
	return persons, nil
}

func (r *Repository) FindOrCreatePerson(userID int64, name string) (*models.Person, error) {
	var person models.Person
	query := `SELECT id, name, description, created_by_user_id, created_at, updated_at FROM persons WHERE created_by_user_id = ? AND name = ?`
	if err := r.DB.Get(&person, query, userID, name); err == nil {
		return &person, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	person = models.Person{Name: name, CreatedBy: userID}
	if err := r.CreatePerson(&person); err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *Repository) ListHashtags() ([]models.Hashtag, error) {
	var tags []models.Hashtag
	if err := r.DB.Select(&tags, `SELECT id, name, created_at FROM hashtags ORDER BY name ASC`); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *Repository) ListHashtagsByUser(userID int64) ([]models.Hashtag, error) {
	var tags []models.Hashtag
	query := `SELECT DISTINCT h.id, h.name, h.created_at
		FROM hashtags h
		JOIN post_hashtags ph ON ph.hashtag_id = h.id
		JOIN posts p ON p.id = ph.post_id
		WHERE p.user_id = ?
		ORDER BY h.name ASC`
	if err := r.DB.Select(&tags, query, userID); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *Repository) FindOrCreateHashtag(name string) (*models.Hashtag, error) {
	var tag models.Hashtag
	query := `SELECT id, name, created_at FROM hashtags WHERE name = ?`
	if err := r.DB.Get(&tag, query, name); err == nil {
		return &tag, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	res, err := r.DB.Exec(`INSERT INTO hashtags (name, created_at) VALUES (?, NOW())`, name)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	tag.ID = id
	tag.Name = name
	tag.CreatedAt = time.Now()
	return &tag, nil
}

func (r *Repository) SavePostWithRelations(userID int64, post *models.Post, tagNames, personNames []string) error {
	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if post.ID == 0 {
		query := `INSERT INTO posts (user_id, date, text, category, mood, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, NOW(), NOW())`
		res, execErr := tx.Exec(query, post.UserID, post.Date, post.Text, post.Category, post.Mood)
		if execErr != nil {
			err = execErr
			return err
		}
		id, _ := res.LastInsertId()
		post.ID = id
	} else {
		if _, execErr := tx.Exec(`UPDATE posts SET text = ?, category = ?, mood = ?, updated_at = NOW() WHERE id = ? AND user_id = ?`,
			post.Text, post.Category, post.Mood, post.ID, post.UserID); execErr != nil {
			err = execErr
			return err
		}
	}

	var tagModels []models.Hashtag
	for _, tag := range tagNames {
		model, execErr := findOrCreateHashtagTx(tx, tag)
		if execErr != nil {
			err = execErr
			return err
		}
		tagModels = append(tagModels, *model)
	}

	var personModels []models.Person
	for _, name := range personNames {
		model, execErr := findOrCreatePersonTx(tx, userID, name)
		if execErr != nil {
			err = execErr
			return err
		}
		personModels = append(personModels, *model)
	}

	if execErr := replacePostTagsTx(tx, post.ID, tagModels); execErr != nil {
		err = execErr
		return err
	}
	if execErr := replacePostMentionsTx(tx, post.ID, personModels); execErr != nil {
		err = execErr
		return err
	}

	if execErr := tx.Commit(); execErr != nil {
		err = execErr
		return err
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
	id, _ := res.LastInsertId()
	post.ID = id
	return nil
}

func (r *Repository) UpdatePost(post *models.Post) error {
	_, err := r.DB.Exec(`UPDATE posts SET text = ?, category = ?, mood = ?, updated_at = NOW() WHERE id = ? AND user_id = ?`,
		post.Text, post.Category, post.Mood, post.ID, post.UserID)
	return err
}

func (r *Repository) DeletePost(id, userID int64) error {
	_, err := r.DB.Exec(`DELETE FROM posts WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *Repository) GetPost(id, userID int64) (*models.Post, error) {
	var post models.Post
	query := `SELECT id, user_id, date, text, category, mood, created_at, updated_at FROM posts WHERE id = ? AND user_id = ?`
	if err := r.DB.Get(&post, query, id, userID); err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *Repository) ListPosts(userID int64, date time.Time, hashtags, persons []string, search string) ([]models.Post, error) {
	base := `SELECT DISTINCT p.id, p.user_id, p.date, p.text, p.category, p.mood, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN post_hashtags ph ON ph.post_id = p.id
		LEFT JOIN hashtags h ON h.id = ph.hashtag_id
		LEFT JOIN mentions m ON m.post_id = p.id
		LEFT JOIN persons pe ON pe.id = m.person_id
		WHERE p.user_id = ? AND p.date = ?`
	args := []interface{}{userID, date}
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

func (r *Repository) CreateComment(comment *models.Comment) error {
	res, err := r.DB.Exec(`INSERT INTO comments (post_id, user_id, text, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`, comment.PostID, comment.UserID, comment.Text)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	comment.ID = id
	return nil
}

func (r *Repository) UpdateComment(comment *models.Comment) error {
	_, err := r.DB.Exec(`UPDATE comments SET text = ?, updated_at = NOW() WHERE id = ? AND user_id = ?`, comment.Text, comment.ID, comment.UserID)
	return err
}

func (r *Repository) DeleteComment(id, userID int64) error {
	_, err := r.DB.Exec(`DELETE FROM comments WHERE id = ? AND user_id = ?`, id, userID)
	return err
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

func (r *Repository) CreateAttachment(att *models.Attachment) error {
	res, err := r.DB.Exec(`INSERT INTO attachments (post_id, file_name, file_type, file_size, url, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())`, att.PostID, att.FileName, att.FileType, att.FileSize, att.URL)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	att.ID = id
	return nil
}

func (r *Repository) ListTagsForPosts(postIDs []int64) (map[int64][]models.Hashtag, error) {
	results := make(map[int64][]models.Hashtag)
	if len(postIDs) == 0 {
		return results, nil
	}
	query, args, err := sqlx.In(`SELECT ph.post_id, h.id, h.name, h.created_at
		FROM post_hashtags ph
		JOIN hashtags h ON h.id = ph.hashtag_id
		WHERE ph.post_id IN (?)`, postIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB.Rebind(query)
	type row struct {
		PostID    int64     `db:"post_id"`
		ID        int64     `db:"id"`
		Name      string    `db:"name"`
		CreatedAt time.Time `db:"created_at"`
	}
	var rows []row
	if err := r.DB.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	for _, item := range rows {
		results[item.PostID] = append(results[item.PostID], models.Hashtag{
			ID:        item.ID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
		})
	}
	return results, nil
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

func findOrCreateHashtagTx(tx *sqlx.Tx, name string) (*models.Hashtag, error) {
	var tag models.Hashtag
	if err := tx.Get(&tag, `SELECT id, name, created_at FROM hashtags WHERE name = ?`, name); err == nil {
		return &tag, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	res, err := tx.Exec(`INSERT INTO hashtags (name, created_at) VALUES (?, NOW())`, name)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	tag.ID = id
	tag.Name = name
	tag.CreatedAt = time.Now()
	return &tag, nil
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
	id, _ := res.LastInsertId()
	person.ID = id
	person.CreatedAt = time.Now()
	person.UpdatedAt = time.Now()
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
