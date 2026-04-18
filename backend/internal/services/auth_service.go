package services

import (
	"log"
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

	go func() {
		if err := email.SendRegistrationPending(s.Email, user.Email); err != nil {
			log.Printf("email: registration pending to %s: %v", user.Email, err)
		}
	}()

	if adminEmails, err := s.Users.GetAdminEmails(); err != nil {
		log.Printf("register: failed to get admin emails for notification: %v", err)
	} else {
		go func() {
			if err := email.SendNewUserNotification(s.Email, adminEmails, user.Email); err != nil {
				log.Printf("email: new user notification: %v", err)
			}
		}()
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
