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

// UpdateUserRequest represents a request to update user info
type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	WorkArea  string `json:"work_area"`
}

// AssignCompanyRequest represents a request to assign a user to a company
type AssignCompanyRequest struct {
	UserID    int64  `json:"user_id" validate:"required"`
	CompanyID int64  `json:"company_id" validate:"required"`
	Role      string `json:"role" validate:"required,oneof=viewer editor admin"`
}

// UpdateCompanyRoleRequest represents a request to update a user's role in a company
type UpdateCompanyRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=viewer editor admin"`
}

// RemoveCompanyRequest represents a request to remove a user from a company
type RemoveCompanyRequest struct {
	UserID    int64 `json:"user_id" validate:"required"`
	CompanyID int64 `json:"company_id" validate:"required"`
}

// AssignPermissionsRequest represents a request to assign permissions to a user
type AssignPermissionsRequest struct {
	UserID      int64    `json:"user_id" validate:"required"`
	Permissions []string `json:"permissions" validate:"required"`
}

// ListUsersRequest represents pagination parameters for listing users
type ListUsersRequest struct {
	Page      int   `json:"page"`
	Size      int   `json:"size"`
	CompanyID int64 `json:"company_id"` // Optional filter by company
}

// SetPasswordRequest represents a request from super admin to set a user's password
// Does not require the current password
type SetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ChangePasswordRequest represents a request from user to change their own password
// Requires the current password for verification
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}
