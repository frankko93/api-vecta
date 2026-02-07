package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
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
	ValidateToken(ctx context.Context, token string) (int64, error)
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

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Create session
	session := &auth.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(SessionDuration),
	}

	err = uc.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	// Get user permissions
	permissions, err := uc.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Build response
	userWithPerms := auth.UserWithPermissions{
		User:        *user,
		Permissions: permissions,
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

	userWithPerms := &auth.UserWithPermissions{
		User:        *user,
		Permissions: permissions,
	}

	return userWithPerms, nil
}

// ValidateToken validates a session token and returns the user ID
func (uc *useCase) ValidateToken(ctx context.Context, token string) (int64, error) {
	session, err := uc.repo.GetSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return 0, ErrInvalidToken
		}
		return 0, err
	}

	return session.UserID, nil
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

// HashPassword hashes a password using argon2id
func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}
