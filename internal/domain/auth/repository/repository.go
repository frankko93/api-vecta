package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/gmhafiz/go8/internal/domain/auth"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrSessionNotFound  = errors.New("session not found")
	ErrDNIAlreadyExists = errors.New("dni already exists")
)

// Repository defines the interface for auth data operations
type Repository interface {
	// User operations
	GetUserByDNI(ctx context.Context, dni string) (*auth.User, error)
	GetUserByID(ctx context.Context, id int64) (*auth.User, error)
	CreateUser(ctx context.Context, user *auth.User) error
	ListUsers(ctx context.Context, page, size int) ([]*auth.User, int, error)

	// Permission operations
	GetUserPermissions(ctx context.Context, userID int64) ([]string, error)
	AssignPermissions(ctx context.Context, userID int64, permissionNames []string) error

	// Session operations
	CreateSession(ctx context.Context, session *auth.Session) error
	GetSessionByToken(ctx context.Context, token string) (*auth.Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userID int64) error
}

type repository struct {
	db *sqlx.DB
}

// New creates a new auth repository
func New(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// GetUserByDNI retrieves a user by DNI
func (r *repository) GetUserByDNI(ctx context.Context, dni string) (*auth.User, error) {
	var user auth.User
	query := `
		SELECT id, first_name, last_name, dni, birth_date, work_area, 
		       password_hash, active, created_at, updated_at
		FROM users
		WHERE dni = $1 AND active = true
	`

	err := r.db.GetContext(ctx, &user, query, dni)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *repository) GetUserByID(ctx context.Context, id int64) (*auth.User, error) {
	var user auth.User
	query := `
		SELECT id, first_name, last_name, dni, birth_date, work_area, 
		       password_hash, active, created_at, updated_at
		FROM users
		WHERE id = $1 AND active = true
	`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user
func (r *repository) CreateUser(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (first_name, last_name, dni, birth_date, work_area, password_hash, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.FirstName,
		user.LastName,
		user.DNI,
		user.BirthDate,
		user.WorkArea,
		user.PasswordHash,
		user.Active,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation on DNI
		if err.Error() == `pq: duplicate key value violates unique constraint "users_dni_key"` {
			return ErrDNIAlreadyExists
		}
		return err
	}

	return nil
}

// GetUserPermissions retrieves all permission names for a user
func (r *repository) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	var permissions []string
	query := `
		SELECT p.name
		FROM permissions p
		INNER JOIN user_permissions up ON p.id = up.permission_id
		WHERE up.user_id = $1
		ORDER BY p.name
	`

	err := r.db.SelectContext(ctx, &permissions, query, userID)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// AssignPermissions assigns permissions to a user
func (r *repository) AssignPermissions(ctx context.Context, userID int64, permissionNames []string) error {
	if len(permissionNames) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get permission IDs from names
	query := `SELECT id FROM permissions WHERE name = ANY($1)`
	var permissionIDs []int
	err = tx.SelectContext(ctx, &permissionIDs, query, permissionNames)
	if err != nil {
		return err
	}

	// Insert user permissions
	insertQuery := `
		INSERT INTO user_permissions (user_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, permission_id) DO NOTHING
	`

	for _, permID := range permissionIDs {
		_, err = tx.ExecContext(ctx, insertQuery, userID, permID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CreateSession creates a new session
func (r *repository) CreateSession(ctx context.Context, session *auth.Session) error {
	query := `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, session.Token, session.UserID, session.ExpiresAt)
	return err
}

// GetSessionByToken retrieves a session by token
func (r *repository) GetSessionByToken(ctx context.Context, token string) (*auth.Session, error) {
	var session auth.Session
	query := `
		SELECT token, user_id, expires_at, created_at
		FROM sessions
		WHERE token = $1 AND expires_at > NOW()
	`

	err := r.db.GetContext(ctx, &session, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}

// DeleteSession deletes a session by token
func (r *repository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// DeleteUserSessions deletes all sessions for a user
func (r *repository) DeleteUserSessions(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// ListUsers retrieves paginated list of users
func (r *repository) ListUsers(ctx context.Context, page, size int) ([]*auth.User, int, error) {
	offset := (page - 1) * size

	var users []*auth.User
	query := `
		SELECT id, first_name, last_name, dni, birth_date, work_area, 
		       password_hash, active, created_at, updated_at
		FROM users
		WHERE active = true
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	err := r.db.SelectContext(ctx, &users, query, size, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE active = true`
	err = r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
