package services

import "familyjournal/backend/internal/models"

type AccessScope struct {
	UserID int64
	Role   string
}

func NewAccessScope(userID int64, role string) AccessScope {
	return AccessScope{UserID: userID, Role: role}
}

func (a AccessScope) IsAdmin() bool {
	return a.Role == models.RoleAdmin
}

func (a AccessScope) OwnerFilter() *int64 {
	if a.IsAdmin() {
		return nil
	}
	return &a.UserID
}
