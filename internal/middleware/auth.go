package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gmhafiz/go8/internal/domain/auth/repository"
	"github.com/gmhafiz/go8/internal/domain/auth/usecase"
	"github.com/gmhafiz/go8/internal/utility/respond"
)

type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
)

// RequireAuth is a middleware that validates the session token
func RequireAuth(authUseCase usecase.UseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				respond.Error(w, http.StatusUnauthorized, errors.New("missing authorization token"))
				return
			}

			userID, err := authUseCase.ValidateToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, usecase.ErrInvalidToken) {
					respond.Error(w, http.StatusUnauthorized, err)
					return
				}
				respond.Error(w, http.StatusInternalServerError, err)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
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

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
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
