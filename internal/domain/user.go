package domain

import "time"

type User struct {
	ID        string
	Email     string
	Telegram  string
	Role      UserRole
	CreatedAt time.Time
}

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)
