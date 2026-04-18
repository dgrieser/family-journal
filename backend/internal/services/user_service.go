package services

import (
	"database/sql"
	"errors"

	"familyjournal/backend/internal/email"
	"familyjournal/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) GetUserByID(userID int64) (*models.User, error) {
	return s.Users.GetUserByID(userID)
}

func (s *Service) UpdateUserProfile(userID int64, email string) error {
	return s.Users.UpdateUserProfile(userID, email)
}

func (s *Service) ChangePassword(userID int64, currentPassword, newPassword string) error {
	if len(newPassword) < 6 || len(newPassword) > 72 {
		return ErrInvalidPassword
	}

	user, err := s.Users.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Users.UpdateUserPassword(userID, string(hash))
}

func (s *Service) ListUsers() ([]models.User, error) {
	users, err := s.Users.ListUsers()
	if err != nil {
		return nil, err
	}
	return ensureSlice(users), nil
}

func (s *Service) UpdateUserRole(userID int64, role string) error {
	return s.Users.UpdateUserRole(userID, role)
}

func (s *Service) UpdateUserActive(userID int64, active bool) (*models.User, error) {
	user, err := s.Users.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user.IsActive == active {
		return user, nil
	}
	if err := s.Users.UpdateUserActive(userID, active); err != nil {
		return nil, err
	}
	user.IsActive = active
	if active {
		email.SendAccountActivated(s.Email, user.Email)
	}
	return user, nil
}
