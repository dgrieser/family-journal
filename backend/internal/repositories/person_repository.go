package repositories

import (
	"database/sql"

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

func (r *Repository) ListPersons(ownerFilter *int64, limit, offset int) ([]models.Person, error) {
	var persons []models.Person
	query := `SELECT id, name, description, created_by_user_id, created_at, updated_at FROM persons`
	args := []interface{}{}
	if ownerFilter != nil {
		query += ` WHERE created_by_user_id = ?`
		args = append(args, *ownerFilter)
	}
	query += ` ORDER BY name ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	if err := r.DB.Select(&persons, query, args...); err != nil {
		return nil, err
	}
	return persons, nil
}

func (r *Repository) CountPersons(ownerFilter *int64) (int, error) {
	var total int
	query := `SELECT COUNT(*) FROM persons`
	args := []interface{}{}
	if ownerFilter != nil {
		query += ` WHERE created_by_user_id = ?`
		args = append(args, *ownerFilter)
	}
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
