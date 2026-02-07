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
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
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
	if err := r.DB.Get(&person, `SELECT id, name, description, created_by_user_id, created_at, updated_at FROM persons WHERE id = ?`, person.ID); err != nil {
		return nil, err
	}
	return &person, nil
}
