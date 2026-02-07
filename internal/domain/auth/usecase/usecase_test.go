package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gmhafiz/go8/internal/domain/auth"
	"github.com/gmhafiz/go8/internal/domain/auth/repository"
)

// MockRepository is a mock implementation of auth.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetUserByDNI(ctx context.Context, dni string) (*auth.User, error) {
	args := m.Called(ctx, dni)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id int64) (*auth.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRepository) AssignPermissions(ctx context.Context, userID int64, permissionNames []string) error {
	args := m.Called(ctx, userID, permissionNames)
	return args.Error(0)
}

func (m *MockRepository) CreateSession(ctx context.Context, session *auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) GetSessionByToken(ctx context.Context, token string) (*auth.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Session), args.Error(1)
}

func (m *MockRepository) DeleteSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRepository) DeleteUserSessions(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRepository) ListUsers(ctx context.Context, page, size int) ([]*auth.User, int, error) {
	args := m.Called(ctx, page, size)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*auth.User), args.Int(1), args.Error(2)
}

// TestLogin tests the login flow
func TestLogin_Success(t *testing.T) {
	mockRepo, uc := setupUseCase()
	ctx := getTestContext()

	testUser := newTestAdminUser()

	mockRepo.On("GetUserByDNI", ctx, testUser.DNI).Return(testUser, nil)
	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*auth.Session")).Return(nil)
	mockRepo.On("GetUserPermissions", ctx, testUser.ID).Return([]string{"admin"}, nil)

	req := &auth.LoginRequest{
		DNI:      testUser.DNI,
		Password: TestPassword,
	}

	response, err := uc.Login(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, testUser.FirstName, response.User.FirstName)
	assert.Contains(t, response.User.Permissions, "admin")

	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mockRepo, uc := setupUseCase()
	ctx := getTestContext()

	mockRepo.On("GetUserByDNI", ctx, TestDNI).Return(nil, repository.ErrUserNotFound)

	req := &auth.LoginRequest{
		DNI:      TestDNI,
		Password: "wrongpassword",
	}

	response, err := uc.Login(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

func TestValidateToken_Success(t *testing.T) {
	mockRepo, uc := setupUseCase()
	ctx := getTestContext()
	token := "valid_token_123"

	session := newTestSession(token, TestUserID)

	mockRepo.On("GetSessionByToken", ctx, token).Return(session, nil)

	userID, err := uc.ValidateToken(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, TestUserID, userID)

	mockRepo.AssertExpectations(t)
}

func TestValidateToken_Invalid(t *testing.T) {
	mockRepo, uc := setupUseCase()
	ctx := getTestContext()
	token := "invalid_token"

	mockRepo.On("GetSessionByToken", ctx, token).Return(nil, repository.ErrSessionNotFound)

	userID, err := uc.ValidateToken(ctx, token)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Equal(t, int64(0), userID)

	mockRepo.AssertExpectations(t)
}
