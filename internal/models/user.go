package models

import (
	"time"
)

type UserRole string

const (
	RoleUser       UserRole = "user"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "superadmin"
)

type User struct {
	ID         int64
	TelegramID int64
	Username   *string
	Phone      *string
	Email      *string
	CreatedAt  time.Time
	Role       UserRole
	IsBanned   bool
}
