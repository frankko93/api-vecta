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

// UserDetailResponse includes user data with their company associations
type UserDetailResponse struct {
	ID          int64         `json:"id"`
	FirstName   string        `json:"first_name"`
	LastName    string        `json:"last_name"`
	DNI         string        `json:"dni"`
	BirthDate   time.Time     `json:"birth_date"`
	WorkArea    string        `json:"work_area"`
	Active      bool          `json:"active"`
	Permissions []string      `json:"permissions"`
	Companies   []UserCompany `json:"companies"`
	CreatedAt   time.Time     `json:"created_at"`
}

// UsersListResponse represents a paginated list of users
type UsersListResponse struct {
	Users      []UserDetailResponse `json:"users"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	Size       int                  `json:"size"`
	TotalPages int                  `json:"total_pages"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

