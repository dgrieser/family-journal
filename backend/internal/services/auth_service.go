package services

import (
	"familyjournal/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(email, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Email:    email,
		Password: string(hash),
		Role:     models.RoleUser,
		IsActive: true,
	}
	if err := s.Users.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Authenticate(email, password string) (*models.User, error) {
	user, err := s.Users.GetUserByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrInactiveUser
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
