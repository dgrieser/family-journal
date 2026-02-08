package repositories

import (
	"database/sql"
	"strings"
	"time"

	"familyjournal/backend/internal/models"

	"github.com/jmoiron/sqlx"
)

func (r *Repository) CreateHashtag(tag *models.Hashtag) error {
	res, err := r.DB.Exec(`INSERT INTO hashtags (name, created_at) VALUES (?, NOW())`, tag.Name)
	if err != nil {
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
	return r.DB.Get(tag, `SELECT id, name, created_at FROM hashtags WHERE id = ?`, id)
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
	name = strings.ToLower(name)
	var tag models.Hashtag
	query := `SELECT id, name, created_at FROM hashtags WHERE name = ?`
	if err := r.DB.Get(&tag, query, name); err == nil {
		return &tag, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	tag = models.Hashtag{Name: name}
	if err := r.CreateHashtag(&tag); err != nil {
		if err = resolveDuplicateInsert(err, func() error { return r.DB.Get(&tag, query, name) }); err != nil {
			return nil, err
		}
		return &tag, nil
	}
	return &tag, nil
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

func findOrCreateHashtagTx(tx *sqlx.Tx, name string) (*models.Hashtag, error) {
	name = strings.ToLower(name)
	var tag models.Hashtag
	if err := tx.Get(&tag, `SELECT id, name, created_at FROM hashtags WHERE name = ?`, name); err == nil {
		return &tag, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	tag = models.Hashtag{Name: name}
	if err := createHashtagTx(tx, &tag); err != nil {
		if err = resolveDuplicateInsert(err, func() error {
			return tx.Get(&tag, `SELECT id, name, created_at FROM hashtags WHERE name = ?`, name)
		}); err != nil {
			return nil, err
		}
		return &tag, nil
	}
	return &tag, nil
}

func createHashtagTx(tx *sqlx.Tx, tag *models.Hashtag) error {
	res, err := tx.Exec(`INSERT INTO hashtags (name, created_at) VALUES (?, NOW())`, tag.Name)
	if err != nil {
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
	return tx.Get(tag, `SELECT id, name, created_at FROM hashtags WHERE id = ?`, id)
}
