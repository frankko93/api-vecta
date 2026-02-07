package auth

import "time"

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token string              `json:"token"`
	User  UserWithPermissions `json:"user"`
}

// UserResponse represents user data for API responses
type UserResponse struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DNI         string    `json:"dni"`
	BirthDate   time.Time `json:"birth_date"`
	WorkArea    string    `json:"work_area"`
	Active      bool      `json:"active"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

