package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/gmhafiz/go8/internal/domain/auth"
	authRepo "github.com/gmhafiz/go8/internal/domain/auth/repository"
	"github.com/gmhafiz/go8/internal/domain/auth/usecase"
	"github.com/gmhafiz/go8/internal/utility/request"
	"github.com/gmhafiz/go8/internal/utility/respond"
)

type Handler struct {
	useCase   usecase.UseCase
	validator *validator.Validate
	repo      authRepo.Repository
}

// RegisterHTTPEndPoints registers auth HTTP endpoints
func RegisterHTTPEndPoints(router *chi.Mux, validator *validator.Validate, uc usecase.UseCase, repo authRepo.Repository) *Handler {
	h := &Handler{
		useCase:   uc,
		validator: validator,
		repo:      repo,
	}

	router.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/logout", h.Logout)
		r.Get("/me", h.Me)
	})

	return h
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return session token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login credentials"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} respond.Error
// @Failure 401 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.useCase.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			respond.Error(w, http.StatusUnauthorized, err)
			return
		}
		if errors.Is(err, usecase.ErrUserInactive) {
			respond.Error(w, http.StatusForbidden, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, response)
}

// Logout handles user logout
// @Summary User logout
// @Description Invalidate current session
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} auth.MessageResponse
// @Failure 401 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		respond.Error(w, http.StatusUnauthorized, errors.New("missing authorization token"))
		return
	}

	err := h.useCase.Logout(r.Context(), token)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, auth.MessageResponse{Message: "logged out successfully"})
}

// Me returns current user information
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} auth.UserWithPermissions
// @Failure 401 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /api/v1/auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		respond.Error(w, http.StatusUnauthorized, errors.New("missing authorization token"))
		return
	}

	user, err := h.useCase.GetCurrentUser(r.Context(), token)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidToken) {
			respond.Error(w, http.StatusUnauthorized, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, user)
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

// ListUsers returns paginated list of users with permissions
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	page := 1
	size := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	users, total, err := h.repo.ListUsers(r.Context(), page, size)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	// Add permissions to each user
	usersWithPermissions := make([]auth.UserWithPermissions, len(users))
	for i, user := range users {
		permissions, err := h.repo.GetUserPermissions(r.Context(), user.ID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, err)
			return
		}

		usersWithPermissions[i] = auth.UserWithPermissions{
			User:        *user,
			Permissions: permissions,
		}
	}

	response := map[string]interface{}{
		"data": usersWithPermissions,
		"pagination": map[string]interface{}{
			"page":  page,
			"size":  size,
			"total": total,
		},
	}

	respond.JSON(w, http.StatusOK, response)
}
