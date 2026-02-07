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
	"github.com/gmhafiz/go8/internal/middleware"
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

// ListUsers returns paginated list of users with permissions and companies
// Can be filtered by company_id query param for company admins
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")
	companyIDStr := r.URL.Query().Get("company_id")

	page := 1
	size := 10
	var companyID int64

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

	if companyIDStr != "" {
		if cid, err := strconv.ParseInt(companyIDStr, 10, 64); err == nil && cid > 0 {
			companyID = cid
		}
	}

	var users []*auth.User
	var total int
	var err error

	if companyID > 0 {
		// List users for a specific company
		users, total, err = h.repo.ListUsersByCompany(r.Context(), companyID, page, size)
	} else {
		// List all users (super admin only)
		users, total, err = h.repo.ListUsers(r.Context(), page, size)
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	// Build detailed response with permissions and companies
	usersResponse := make([]auth.UserDetailResponse, len(users))
	for i, user := range users {
		permissions, err := h.repo.GetUserPermissions(r.Context(), user.ID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, err)
			return
		}

		companies, err := h.repo.GetUserCompanies(r.Context(), user.ID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, err)
			return
		}

		usersResponse[i] = auth.UserDetailResponse{
			ID:          user.ID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			DNI:         user.DNI,
			BirthDate:   user.BirthDate,
			WorkArea:    user.WorkArea,
			Active:      user.Active,
			Permissions: permissions,
			Companies:   companies,
			CreatedAt:   user.CreatedAt,
		}
	}

	totalPages := total / size
	if total%size > 0 {
		totalPages++
	}

	response := auth.UsersListResponse{
		Users:      usersResponse,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	}

	respond.JSON(w, http.StatusOK, response)
}

// GetUser returns a specific user by ID
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid user id"))
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, authRepo.ErrUserNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	permissions, err := h.repo.GetUserPermissions(r.Context(), user.ID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	companies, err := h.repo.GetUserCompanies(r.Context(), user.ID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	response := auth.UserDetailResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DNI:         user.DNI,
		BirthDate:   user.BirthDate,
		WorkArea:    user.WorkArea,
		Active:      user.Active,
		Permissions: permissions,
		Companies:   companies,
		CreatedAt:   user.CreatedAt,
	}

	respond.JSON(w, http.StatusOK, response)
}

// CreateUser creates a new user
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req auth.CreateUserRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.useCase.CreateUser(r.Context(), &req)
	if err != nil {
		if errors.Is(err, authRepo.ErrDNIAlreadyExists) {
			respond.Error(w, http.StatusConflict, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	response := auth.UserDetailResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DNI:         user.DNI,
		BirthDate:   user.BirthDate,
		WorkArea:    user.WorkArea,
		Active:      user.Active,
		Permissions: req.Permissions,
		Companies:   []auth.UserCompany{},
		CreatedAt:   user.CreatedAt,
	}

	respond.JSON(w, http.StatusCreated, response)
}

// UpdateUser updates user information
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid user id"))
		return
	}

	var req auth.UpdateUserRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	// Get existing user
	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, authRepo.ErrUserNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	// Update fields if provided
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.WorkArea != "" {
		user.WorkArea = req.WorkArea
	}

	err = h.repo.UpdateUser(r.Context(), user)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	// Fetch updated data for response
	permissions, _ := h.repo.GetUserPermissions(r.Context(), user.ID)
	companies, _ := h.repo.GetUserCompanies(r.Context(), user.ID)

	response := auth.UserDetailResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DNI:         user.DNI,
		BirthDate:   user.BirthDate,
		WorkArea:    user.WorkArea,
		Active:      user.Active,
		Permissions: permissions,
		Companies:   companies,
		CreatedAt:   user.CreatedAt,
	}

	respond.JSON(w, http.StatusOK, response)
}

// DeactivateUser soft deletes a user
func (h *Handler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid user id"))
		return
	}

	err = h.repo.DeactivateUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, authRepo.ErrUserNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	// Also invalidate all user sessions
	_ = h.repo.DeleteUserSessions(r.Context(), userID)

	respond.JSON(w, http.StatusOK, auth.MessageResponse{Message: "user deactivated successfully"})
}

// AssignUserToCompany assigns a user to a company with a role
func (h *Handler) AssignUserToCompany(w http.ResponseWriter, r *http.Request) {
	var req auth.AssignCompanyRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	// Verify user exists
	_, err := h.repo.GetUserByID(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, authRepo.ErrUserNotFound) {
			respond.Error(w, http.StatusNotFound, errors.New("user not found"))
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	err = h.repo.AssignUserToCompany(r.Context(), req.UserID, req.CompanyID, req.Role)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	// Invalidate user's sessions so they get new company_roles on next login
	_ = h.repo.DeleteUserSessions(r.Context(), req.UserID)

	respond.JSON(w, http.StatusOK, auth.MessageResponse{
		Message: "user assigned to company successfully",
	})
}

// UpdateUserCompanyRole updates a user's role in a company
func (h *Handler) UpdateUserCompanyRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	companyIDStr := chi.URLParam(r, "company_id")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid user id"))
		return
	}

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company id"))
		return
	}

	var req auth.UpdateCompanyRoleRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.repo.UpdateUserCompanyRole(r.Context(), userID, companyID, req.Role)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	// Invalidate user's sessions so they get new company_roles on next login
	_ = h.repo.DeleteUserSessions(r.Context(), userID)

	respond.JSON(w, http.StatusOK, auth.MessageResponse{
		Message: "user role updated successfully",
	})
}

// RemoveUserFromCompany removes a user's access to a company
func (h *Handler) RemoveUserFromCompany(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	companyIDStr := chi.URLParam(r, "company_id")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid user id"))
		return
	}

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company id"))
		return
	}

	err = h.repo.RemoveUserFromCompany(r.Context(), userID, companyID)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	// Invalidate user's sessions so they get updated company_roles on next login
	_ = h.repo.DeleteUserSessions(r.Context(), userID)

	respond.JSON(w, http.StatusOK, auth.MessageResponse{
		Message: "user removed from company successfully",
	})
}

// AssignPermissions assigns permissions to a user
func (h *Handler) AssignPermissions(w http.ResponseWriter, r *http.Request) {
	var req auth.AssignPermissionsRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	// Verify user exists
	_, err := h.repo.GetUserByID(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, authRepo.ErrUserNotFound) {
			respond.Error(w, http.StatusNotFound, errors.New("user not found"))
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	err = h.repo.AssignPermissions(r.Context(), req.UserID, req.Permissions)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, auth.MessageResponse{
		Message: "permissions assigned successfully",
	})
}

// SetPassword allows super admin to set a user's password (no current password required)
func (h *Handler) SetPassword(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid user id"))
		return
	}

	var req auth.SetPasswordRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.useCase.SetPassword(r.Context(), userID, req.NewPassword)
	if err != nil {
		if errors.Is(err, authRepo.ErrUserNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, auth.MessageResponse{
		Message: "password set successfully",
	})
}

// ChangePassword allows authenticated user to change their own password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by RequireAuth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	var req auth.ChangePasswordRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.useCase.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			respond.Error(w, http.StatusUnauthorized, errors.New("current password is incorrect"))
			return
		}
		if errors.Is(err, authRepo.ErrUserNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, auth.MessageResponse{
		Message: "password changed successfully - please login again",
	})
}
