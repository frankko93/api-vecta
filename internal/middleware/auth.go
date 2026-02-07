package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gmhafiz/go8/internal/domain/auth"
	"github.com/gmhafiz/go8/internal/domain/auth/repository"
	"github.com/gmhafiz/go8/internal/domain/auth/usecase"
	"github.com/gmhafiz/go8/internal/utility/respond"
)

type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// CompanyRolesKey is the context key for all user's company roles (from session)
	CompanyRolesKey contextKey = "company_roles"
	// CompanyIDKey is the context key for the current request's company ID
	CompanyIDKey contextKey = "company_id"
	// CompanyRoleKey is the context key for user's role in the current company
	CompanyRoleKey contextKey = "company_role"
)

// =============================================================================
// Company Roles - Define what each role can do within a company
// =============================================================================
// Role hierarchy: admin > editor > viewer
// Higher roles inherit all permissions from lower roles

type CompanyRole string

const (
	// RoleViewer can only read data (reports, summaries, lists)
	RoleViewer CompanyRole = "viewer"

	// RoleEditor can read + write data (import CSV, save reports)
	RoleEditor CompanyRole = "editor"

	// RoleAdmin can read + write + delete + manage (full control over company data)
	RoleAdmin CompanyRole = "admin"
)

// RoleLevel returns the numeric level of a role for comparison
// Higher level = more permissions
func (r CompanyRole) Level() int {
	switch r {
	case RoleViewer:
		return 1
	case RoleEditor:
		return 2
	case RoleAdmin:
		return 3
	default:
		return 0
	}
}

// HasAtLeast checks if this role has at least the permissions of the required role
func (r CompanyRole) HasAtLeast(required CompanyRole) bool {
	return r.Level() >= required.Level()
}

// IsValid checks if the role is a valid company role
func (r CompanyRole) IsValid() bool {
	return r.Level() > 0
}

var (
	ErrCompanyAccessDenied    = errors.New("you don't have access to this company")
	ErrInsufficientRole       = errors.New("insufficient role for this action")
	ErrMissingCompanyID       = errors.New("missing or invalid company_id")
	ErrInvalidRole            = errors.New("invalid company role")
)

// RequireAuth is a middleware that validates the session token
// It stores user ID and company roles in context for use by other middlewares
func RequireAuth(authUseCase usecase.UseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				respond.Error(w, http.StatusUnauthorized, errors.New("missing authorization token"))
				return
			}

			session, err := authUseCase.ValidateToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, usecase.ErrInvalidToken) {
					respond.Error(w, http.StatusUnauthorized, err)
					return
				}
				respond.Error(w, http.StatusInternalServerError, err)
				return
			}

			// Add user ID and company roles to context (from session cache)
			ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
			ctx = context.WithValue(ctx, CompanyRolesKey, session.CompanyRoles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission is a middleware that validates if user has a specific permission
func RequirePermission(authRepo repository.Repository, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok {
				respond.Error(w, http.StatusUnauthorized, errors.New("user not authenticated"))
				return
			}

			permissions, err := authRepo.GetUserPermissions(r.Context(), userID)
			if err != nil {
				respond.Error(w, http.StatusInternalServerError, err)
				return
			}

			hasPermission := false
			for _, p := range permissions {
				if p == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				respond.Error(w, http.StatusForbidden, fmt.Errorf("insufficient permissions: requires '%s'", permission))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateCompanyAccess is a middleware that validates if user has access to the requested company
// It reads the role from session cache (context) - NO database query needed
// MUST be used AFTER RequireAuth middleware
func ValidateCompanyAccess(_ repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get company roles from context (set by RequireAuth)
			companyRoles, ok := GetCompanyRoles(r.Context())
			if !ok {
				respond.Error(w, http.StatusUnauthorized, errors.New("user not authenticated"))
				return
			}

			// Try to get company_id from query params first, then from form data
			companyIDStr := r.URL.Query().Get("company_id")
			if companyIDStr == "" {
				// Try form value (for POST requests)
				companyIDStr = r.FormValue("company_id")
			}

			if companyIDStr == "" {
				respond.Error(w, http.StatusBadRequest, ErrMissingCompanyID)
				return
			}

			companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
			if err != nil || companyID <= 0 {
				respond.Error(w, http.StatusBadRequest, ErrMissingCompanyID)
				return
			}

			// Get user's role in this company from session cache (no DB query!)
			roleStr := companyRoles.GetRole(companyID)
			if roleStr == "" {
				respond.Error(w, http.StatusForbidden, ErrCompanyAccessDenied)
				return
			}

			role := CompanyRole(roleStr)
			if !role.IsValid() {
				respond.Error(w, http.StatusInternalServerError, ErrInvalidRole)
				return
			}

			// Add company ID and role to context for downstream handlers/middlewares
			ctx := context.WithValue(r.Context(), CompanyIDKey, companyID)
			ctx = context.WithValue(ctx, CompanyRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireCompanyRole is a middleware that validates if user has at least the required role
// MUST be used AFTER ValidateCompanyAccess middleware
func RequireCompanyRole(requiredRole CompanyRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := GetCompanyRole(r.Context())
			if !ok {
				respond.Error(w, http.StatusInternalServerError, errors.New("company role not found in context - ensure ValidateCompanyAccess runs first"))
				return
			}

			if !role.HasAtLeast(requiredRole) {
				respond.Error(w, http.StatusForbidden, fmt.Errorf("%w: requires '%s' role, you have '%s'", ErrInsufficientRole, requiredRole, role))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetCompanyID extracts the validated company ID from the request context
func GetCompanyID(ctx context.Context) (int64, bool) {
	companyID, ok := ctx.Value(CompanyIDKey).(int64)
	return companyID, ok
}

// GetCompanyRole extracts the user's role in the company from the request context
func GetCompanyRole(ctx context.Context) (CompanyRole, bool) {
	role, ok := ctx.Value(CompanyRoleKey).(CompanyRole)
	return role, ok
}

// GetCompanyRoles extracts all company roles from the request context (from session cache)
func GetCompanyRoles(ctx context.Context) (auth.CompanyRoles, bool) {
	roles, ok := ctx.Value(CompanyRolesKey).(auth.CompanyRoles)
	return roles, ok
}

// CheckCompanyAccess is a helper function that handlers can call to validate company access
// Use this for endpoints where company_id comes from JSON body
// Reads from session cache in context - NO database query needed
// Returns the user's role in the company if access is granted
func CheckCompanyAccess(ctx context.Context, companyID int64) (CompanyRole, error) {
	companyRoles, ok := GetCompanyRoles(ctx)
	if !ok {
		return "", errors.New("company roles not found in context - ensure RequireAuth runs first")
	}

	roleStr := companyRoles.GetRole(companyID)
	if roleStr == "" {
		return "", ErrCompanyAccessDenied
	}

	role := CompanyRole(roleStr)
	if !role.IsValid() {
		return "", ErrInvalidRole
	}

	return role, nil
}

// CheckCompanyRole is a helper function to validate that a user has the required role for a company
// Use this for endpoints where company_id comes from JSON body
// Reads from session cache in context - NO database query needed
func CheckCompanyRole(ctx context.Context, companyID int64, requiredRole CompanyRole) error {
	role, err := CheckCompanyAccess(ctx, companyID)
	if err != nil {
		return err
	}

	if !role.HasAtLeast(requiredRole) {
		return fmt.Errorf("%w: requires '%s' role, you have '%s'", ErrInsufficientRole, requiredRole, role)
	}

	return nil
}

// extractToken extracts the Bearer token from Authorization header
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
