package models

import (
	"time"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Role         UserRole   `gorm:"type:varchar(20);default:'user';not null" json:"role"`
	IsActive     bool       `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
