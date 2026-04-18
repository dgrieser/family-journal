package services

import (
	"net/mail"
	"strings"

	"familyjournal/backend/internal/email"
	"familyjournal/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(userEmail, password string) (*models.User, error) {
	userEmail = strings.ToLower(strings.TrimSpace(userEmail))
	addr, err := mail.ParseAddress(userEmail)
	if err != nil || addr.Address != userEmail {
		return nil, ErrInvalidEmail
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Email:    userEmail,
		Password: string(hash),
		Role:     models.RoleUser,
		IsActive: false,
	}
	if err := s.Users.CreateUser(user); err != nil {
		return nil, err
	}

	email.SendRegistrationPending(s.Email, user.Email)

	if adminEmails, err := s.Users.GetAdminEmails(); err == nil {
		email.SendNewUserNotification(s.Email, adminEmails, user.Email)
	}

	return user, nil
}

func (s *Service) Authenticate(userEmail, password string) (*models.User, error) {
	user, err := s.Users.GetUserByEmail(userEmail)
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
