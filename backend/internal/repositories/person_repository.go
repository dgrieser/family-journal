package repositories

import (
	"database/sql"
	"strings"

	"familyjournal/backend/internal/models"
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
	return r.DB.Get(person, `SELECT id, name, description, created_by_user_id, created_at, updated_at FROM persons WHERE id = ?`, id)
}

func (r *Repository) UpdatePerson(person *models.Person, ownerFilter *int64) error {
	query := `UPDATE persons SET name = ?, description = ?, updated_at = NOW() WHERE id = ?`
	args := []interface{}{person.Name, person.Description, person.ID}
	if ownerFilter != nil {
		query += ` AND created_by_user_id = ?`
		args = append(args, *ownerFilter)
	}
	_, err := r.DB.Exec(query, args...)
	if err != nil && isDuplicateKeyError(err) {
		return models.ErrDuplicate
	}
	return err
}

func (r *Repository) DeletePerson(id int64, ownerFilter *int64) error {
	query := `DELETE FROM persons WHERE id = ?`
	args := []interface{}{id}
	if ownerFilter != nil {
		query += ` AND created_by_user_id = ?`
		args = append(args, *ownerFilter)
	}
	_, err := r.DB.Exec(query, args...)
	return err
}

func buildPersonQuery(ownerFilter *int64, search string) (string, []interface{}) {
	var conditions []string
	args := []interface{}{}

	if ownerFilter != nil {
		conditions = append(conditions, `created_by_user_id = ?`)
		args = append(args, *ownerFilter)
	}
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		conditions = append(conditions, `name LIKE ?`)
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
	query := `SELECT id, name, description, created_by_user_id, created_at, updated_at FROM persons` + whereClause
	query += ` ORDER BY name ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	if err := r.DB.Select(&persons, query, args...); err != nil {
		return nil, err
	}
	return persons, nil
}

func (r *Repository) CountPersons(ownerFilter *int64, search string) (int, error) {
	var total int
	whereClause, args := buildPersonQuery(ownerFilter, search)
	query := `SELECT COUNT(*) FROM persons` + whereClause
	if err := r.DB.Get(&total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
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
		if err = resolveDuplicateInsert(err, func() error { return r.DB.Get(&person, query, userID, name) }); err != nil {
			return nil, err
		}
		return &person, nil
	}
	return &person, nil
}
