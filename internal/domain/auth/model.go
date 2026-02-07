package auth

import (
	"fmt"
	"time"
)

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

// CompanyRoles maps company_id (as string) to role
// Example: {"1": "admin", "2": "viewer"}
type CompanyRoles map[string]string

// Session represents an active user session
type Session struct {
	Token        string       `db:"token" json:"token"`
	UserID       int64        `db:"user_id" json:"user_id"`
	CompanyRoles CompanyRoles `db:"company_roles" json:"company_roles"` // Cached roles per company
	ExpiresAt    time.Time    `db:"expires_at" json:"expires_at"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
}

// GetRole returns the user's role for a specific company
// Returns empty string if user has no access to the company
func (cr CompanyRoles) GetRole(companyID int64) string {
	if cr == nil {
		return ""
	}
	return cr[fmt.Sprintf("%d", companyID)]
}

// HasAccess checks if the user has any role in the specified company
func (cr CompanyRoles) HasAccess(companyID int64) bool {
	return cr.GetRole(companyID) != ""
}

// UserCompany represents a user's association with a company
type UserCompany struct {
	CompanyID   int64  `db:"company_id" json:"company_id"`
	CompanyName string `db:"company_name" json:"company_name"`
	Role        string `db:"role" json:"role"`
}

// UserWithPermissions represents a user with their permissions and companies
type UserWithPermissions struct {
	User
	Permissions []string      `json:"permissions"`
	Companies   []UserCompany `json:"companies"`
}
