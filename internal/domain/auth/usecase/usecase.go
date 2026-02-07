package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"

	"github.com/gmhafiz/go8/internal/domain/auth"
	"github.com/gmhafiz/go8/internal/domain/auth/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserInactive       = errors.New("user is inactive")
)

const (
	// SessionDuration is 6 hours
	SessionDuration = 6 * time.Hour

	// TokenLength in bytes (will be hex encoded, so 32 bytes = 64 chars)
	TokenLength = 32
)

// UseCase defines the interface for auth business logic
type UseCase interface {
	Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error)
	Logout(ctx context.Context, token string) error
	GetCurrentUser(ctx context.Context, token string) (*auth.UserWithPermissions, error)
	ValidateToken(ctx context.Context, token string) (*auth.Session, error)
	CreateUser(ctx context.Context, req *auth.CreateUserRequest) (*auth.User, error)
	SetPassword(ctx context.Context, userID int64, newPassword string) error
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
}

type useCase struct {
	repo repository.Repository
}

// New creates a new auth use case
func New(repo repository.Repository) UseCase {
	return &useCase{repo: repo}
}

// Login authenticates a user and creates a session
func (uc *useCase) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// Get user by DNI
	user, err := uc.repo.GetUserByDNI(ctx, req.DNI)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.Active {
		return nil, ErrUserInactive
	}

	// Verify password
	match, err := argon2id.ComparePasswordAndHash(req.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, ErrInvalidCredentials
	}

	// Get user permissions
	permissions, err := uc.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Get user companies
	companies, err := uc.repo.GetUserCompanies(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Build company roles map for session cache
	companyRoles := make(auth.CompanyRoles)
	for _, c := range companies {
		companyRoles[fmt.Sprintf("%d", c.CompanyID)] = c.Role
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Create session with cached company roles
	session := &auth.Session{
		Token:        token,
		UserID:       user.ID,
		CompanyRoles: companyRoles,
		ExpiresAt:    time.Now().Add(SessionDuration),
	}

	err = uc.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	// Build response
	userWithPerms := auth.UserWithPermissions{
		User:        *user,
		Permissions: permissions,
		Companies:   companies,
	}

	response := &auth.LoginResponse{
		Token: token,
		User:  userWithPerms,
	}

	return response, nil
}

// Logout invalidates a session
func (uc *useCase) Logout(ctx context.Context, token string) error {
	err := uc.repo.DeleteSession(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			// Already logged out, not an error
			return nil
		}
		return err
	}

	return nil
}

// GetCurrentUser retrieves the current user from a session token
func (uc *useCase) GetCurrentUser(ctx context.Context, token string) (*auth.UserWithPermissions, error) {
	// Validate session
	session, err := uc.repo.GetSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	// Get user
	user, err := uc.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	// Get permissions
	permissions, err := uc.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Get companies
	companies, err := uc.repo.GetUserCompanies(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	userWithPerms := &auth.UserWithPermissions{
		User:        *user,
		Permissions: permissions,
		Companies:   companies,
	}

	return userWithPerms, nil
}

// ValidateToken validates a session token and returns the session with company roles
func (uc *useCase) ValidateToken(ctx context.Context, token string) (*auth.Session, error) {
	session, err := uc.repo.GetSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	return session, nil
}

// generateToken creates a cryptographically secure random token
func generateToken() (string, error) {
	bytes := make([]byte, TokenLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// CreateUser creates a new user with the given request data
func (uc *useCase) CreateUser(ctx context.Context, req *auth.CreateUserRequest) (*auth.User, error) {
	// Check if DNI already exists
	existing, err := uc.repo.GetUserByDNI(ctx, req.DNI)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, repository.ErrDNIAlreadyExists
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &auth.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DNI:          req.DNI,
		BirthDate:    req.BirthDate,
		WorkArea:     req.WorkArea,
		PasswordHash: passwordHash,
		Active:       true,
	}

	err = uc.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Assign permissions if provided
	if len(req.Permissions) > 0 {
		err = uc.repo.AssignPermissions(ctx, user.ID, req.Permissions)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

// SetPassword sets a new password for a user (admin action, no current password required)
func (uc *useCase) SetPassword(ctx context.Context, userID int64, newPassword string) error {
	// Verify user exists
	_, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Hash new password
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	err = uc.repo.UpdateUserPassword(ctx, userID, passwordHash)
	if err != nil {
		return err
	}

	// Invalidate all user sessions (force re-login with new password)
	_ = uc.repo.DeleteUserSessions(ctx, userID)

	return nil
}

// ChangePassword allows a user to change their own password (requires current password)
func (uc *useCase) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	// Get user to verify current password
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	match, err := argon2id.ComparePasswordAndHash(currentPassword, user.PasswordHash)
	if err != nil {
		return err
	}

	if !match {
		return ErrInvalidCredentials
	}

	// Hash new password
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	err = uc.repo.UpdateUserPassword(ctx, userID, passwordHash)
	if err != nil {
		return err
	}

	// Invalidate all other sessions (keep current one? or force re-login?)
	// For security, invalidate all sessions including current
	_ = uc.repo.DeleteUserSessions(ctx, userID)

	return nil
}

// HashPassword hashes a password using argon2id
func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}
