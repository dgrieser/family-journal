package repositories

import (
	"database/sql"
	"strings"

	"familyjournal/backend/internal/models"

	"github.com/jmoiron/sqlx"
)

func (r *Repository) CreatePerson(person *models.Person) error {
	query := `INSERT INTO persons (name, description, created_by_user_id, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())`
	res, err := r.DB.Exec(query, person.Name, person.Description, person.CreatedBy)
	if err != nil {
		if isDuplicateKeyError(err) {
			return models.ErrDuplicate
		}
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
	if err := r.DB.Get(person, `SELECT p.id, p.name, p.description, p.created_by_user_id, p.created_at, p.updated_at, u.email AS creator_email FROM persons p JOIN users u ON u.id = p.created_by_user_id WHERE p.id = ?`, id); err != nil {
		return err
	}
	person.HydrateCreator()
	return nil
}

func (r *Repository) UpdatePerson(person *models.Person, ownerFilter *int64) error {
	query := `UPDATE persons SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`
	args := []interface{}{person.Name, person.Description, person.ID}
	if ownerFilter != nil {
		query += ` AND created_by_user_id = ?`
		args = append(args, *ownerFilter)
	}
	res, err := r.DB.Exec(query, args...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return models.ErrDuplicate
		}
		return err
	}
	if ownerFilter != nil {
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return models.ErrForbidden
		}
	}
	return nil
}

func (r *Repository) DeletePerson(id int64, ownerFilter *int64) error {
	query := `DELETE FROM persons WHERE id = ?`
	args := []interface{}{id}
	if ownerFilter != nil {
		query += ` AND created_by_user_id = ?`
		args = append(args, *ownerFilter)
	}
	res, err := r.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	if ownerFilter != nil {
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return models.ErrForbidden
		}
	}
	return nil
}

func buildPersonQuery(ownerFilter *int64, search string) (string, []interface{}) {
	var conditions []string
	args := []interface{}{}

	if ownerFilter != nil {
		conditions = append(conditions, `p.created_by_user_id = ?`)
		args = append(args, *ownerFilter)
	}
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		conditions = append(conditions, `p.name LIKE ?`)
		args = append(args, "%"+trimmed+"%")
	}
	if len(conditions) == 0 {
		return "", args
	}

	return ` WHERE ` + strings.Join(conditions, ` AND `), args
}

func (r *Repository) ListPersons(ownerFilter *int64, search string, limit, offset int) ([]models.Person, error) {
	var persons []models.Person
	whereClause, args := buildPersonQuery(ownerFilter, search)
	query := `SELECT p.id, p.name, p.description, p.created_by_user_id, p.created_at, p.updated_at, u.email AS creator_email FROM persons p JOIN users u ON u.id = p.created_by_user_id` + whereClause
	query += ` ORDER BY p.name ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	if err := r.DB.Select(&persons, query, args...); err != nil {
		return nil, err
	}
	for i := range persons {
		persons[i].HydrateCreator()
	}
	return persons, nil
}

func (r *Repository) ListPersonsPaginated(ownerFilter *int64, search string, limit, offset int) ([]models.Person, int, error) {
	tx, err := r.beginReadSnapshotTx()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	total, err := r.countPersons(tx, ownerFilter, search)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		if err := tx.Commit(); err != nil {
			return nil, 0, err
		}
		return []models.Person{}, 0, nil
	}
	persons, err := r.listPersons(tx, ownerFilter, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return persons, total, nil
}

func (r *Repository) CountPersons(ownerFilter *int64, search string) (int, error) {
	return r.countPersons(r.DB, ownerFilter, search)
}

func (r *Repository) listPersons(queryer sqlx.Ext, ownerFilter *int64, search string, limit, offset int) ([]models.Person, error) {
	var persons []models.Person
	whereClause, args := buildPersonQuery(ownerFilter, search)
	query := `SELECT p.id, p.name, p.description, p.created_by_user_id, p.created_at, p.updated_at, u.email AS creator_email FROM persons p JOIN users u ON u.id = p.created_by_user_id` + whereClause
	query += ` ORDER BY p.name ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	if err := sqlx.Select(queryer, &persons, query, args...); err != nil {
		return nil, err
	}
	for i := range persons {
		persons[i].HydrateCreator()
	}
	return persons, nil
}

func (r *Repository) countPersons(queryer sqlx.Queryer, ownerFilter *int64, search string) (int, error) {
	var total int
	whereClause, args := buildPersonQuery(ownerFilter, search)
	query := `SELECT COUNT(*) FROM persons p` + whereClause
	if err := sqlx.Get(queryer, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) FindOrCreatePerson(userID int64, name string) (*models.Person, error) {
	var person models.Person
	query := `SELECT p.id, p.name, p.description, p.created_by_user_id, p.created_at, p.updated_at, u.email AS creator_email FROM persons p JOIN users u ON u.id = p.created_by_user_id WHERE p.created_by_user_id = ? AND p.name = ?`
	if err := r.DB.Get(&person, query, userID, name); err == nil {
		person.HydrateCreator()
		return &person, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	person = models.Person{Name: name, CreatedBy: userID}
	if err := r.CreatePerson(&person); err != nil {
		if err = resolveDuplicateInsert(err, func() error {
			if err2 := r.DB.Get(&person, query, userID, name); err2 != nil {
				return err2
			}
			person.HydrateCreator()
			return nil
		}); err != nil {
			return nil, err
		}
		return &person, nil
	}
	return &person, nil
}
