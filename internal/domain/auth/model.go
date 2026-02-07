package auth

import "time"

// User represents a user in the system
type User struct {
	ID           int64     `db:"id" json:"id"`
	FirstName    string    `db:"first_name" json:"first_name"`
	LastName     string    `db:"last_name" json:"last_name"`
	DNI          string    `db:"dni" json:"dni"`
	BirthDate    time.Time `db:"birth_date" json:"birth_date"`
	WorkArea     string    `db:"work_area" json:"work_area"`
	PasswordHash string    `db:"password_hash" json:"-"` // Never send to client
	Active       bool      `db:"active" json:"active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          int    `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

// Session represents an active user session
type Session struct {
	Token     string    `db:"token" json:"token"`
	UserID    int64     `db:"user_id" json:"user_id"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// UserWithPermissions represents a user with their permissions
type UserWithPermissions struct {
	User
	Permissions []string `json:"permissions"`
}
