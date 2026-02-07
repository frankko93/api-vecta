package usecase

import (
	"context"
	"time"

	"github.com/gmhafiz/go8/internal/domain/auth"
)

// Test constants - SINGLE SOURCE OF TRUTH for all auth tests
const (
	TestUserID       = int64(1)
	TestDNI          = "99999999"
	TestPassword     = "admin123"
	TestPasswordHash = "$argon2id$v=19$m=65536,t=1,p=11$26wRAe/3D66n2EZzzR0QNw$FLiJupf5T0vQCFLryzB2gWdrR4jLMX8sFVAfq2UbnwE"
)

// newTestAdminUser creates a standard test admin user
// Use this factory in ALL auth tests to maintain consistency
func newTestAdminUser() *auth.User {
	return &auth.User{
		ID:           TestUserID,
		FirstName:    "Admin",
		LastName:     "User",
		DNI:          TestDNI,
		PasswordHash: TestPasswordHash,
		Active:       true,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		WorkArea:     "IT",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// newTestSession creates a standard test session
func newTestSession(token string, userID int64) *auth.Session {
	return &auth.Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(6 * time.Hour),
		CreatedAt: time.Now(),
	}
}

// setupUseCase creates a mock repository and usecase
// Use this in ALL tests to maintain consistent setup
func setupUseCase() (*MockRepository, UseCase) {
	mockRepo := new(MockRepository)
	uc := New(mockRepo)
	return mockRepo, uc
}

// getTestContext returns a standard test context
func getTestContext() context.Context {
	return context.Background()
}
