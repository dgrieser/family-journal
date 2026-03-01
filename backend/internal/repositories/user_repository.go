package repositories

import "familyjournal/backend/internal/models"

func (r *Repository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (email, password_hash, role, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())`
	res, err := r.DB.Exec(query, user.Email, user.Password, user.Role, user.IsActive)
	if err != nil {
		return err
	}
	id, err := lastInsertID(res)
	if err != nil {
		return err
	}
	return r.DB.Get(user, `SELECT id, email, password_hash, role, active, created_at, updated_at FROM users WHERE id = ?`, id)
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

func (r *Repository) UpdateUserPassword(id int64, passwordHash string) error {
	_, err := r.DB.Exec(`UPDATE users SET password_hash = ?, updated_at = NOW() WHERE id = ?`, passwordHash, id)
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
