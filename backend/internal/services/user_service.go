package services

import "familyjournal/backend/internal/models"

func (s *Service) GetUserByID(userID int64) (*models.User, error) {
	return s.Users.GetUserByID(userID)
}

func (s *Service) UpdateUserProfile(userID int64, email string) error {
	return s.Users.UpdateUserProfile(userID, email)
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

func (s *Service) UpdateUserActive(userID int64, active bool) error {
	return s.Users.UpdateUserActive(userID, active)
}
