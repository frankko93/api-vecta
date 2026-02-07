package auth

import "time"

// LoginRequest represents a login request
type LoginRequest struct {
	DNI      string `json:"dni" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	FirstName   string    `json:"first_name" validate:"required"`
	LastName    string    `json:"last_name" validate:"required"`
	DNI         string    `json:"dni" validate:"required"`
	BirthDate   time.Time `json:"birth_date" validate:"required"`
	WorkArea    string    `json:"work_area" validate:"required"`
	Password    string    `json:"password" validate:"required,min=6"`
	Permissions []string  `json:"permissions"` // Optional, defaults to empty
}
