package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
	UpdateUser(ctx context.Context, user *auth.User) error
	UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error
	DeactivateUser(ctx context.Context, userID int64) error
	ListUsers(ctx context.Context, page, size int) ([]*auth.User, int, error)
	ListUsersByCompany(ctx context.Context, companyID int64, page, size int) ([]*auth.User, int, error)

	// Permission operations
	GetUserPermissions(ctx context.Context, userID int64) ([]string, error)
	AssignPermissions(ctx context.Context, userID int64, permissionNames []string) error

	// Company operations
	GetUserCompanies(ctx context.Context, userID int64) ([]auth.UserCompany, error)
	UserHasCompanyAccess(ctx context.Context, userID int64, companyID int64) (bool, error)
	GetUserCompanyRole(ctx context.Context, userID int64, companyID int64) (string, error)
	AssignUserToCompany(ctx context.Context, userID int64, companyID int64, role string) error
	UpdateUserCompanyRole(ctx context.Context, userID int64, companyID int64, role string) error
	RemoveUserFromCompany(ctx context.Context, userID int64, companyID int64) error

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

// CreateSession creates a new session with company roles
func (r *repository) CreateSession(ctx context.Context, session *auth.Session) error {
	// Marshal company roles to JSON
	rolesJSON, err := json.Marshal(session.CompanyRoles)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO sessions (token, user_id, company_roles, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err = r.db.ExecContext(ctx, query, session.Token, session.UserID, rolesJSON, session.ExpiresAt)
	return err
}

// sessionRow is used to scan session data from DB (handles JSONB)
type sessionRow struct {
	Token        string          `db:"token"`
	UserID       int64           `db:"user_id"`
	CompanyRoles json.RawMessage `db:"company_roles"`
	ExpiresAt    sql.NullTime    `db:"expires_at"`
	CreatedAt    sql.NullTime    `db:"created_at"`
}

// GetSessionByToken retrieves a session by token including company roles
func (r *repository) GetSessionByToken(ctx context.Context, token string) (*auth.Session, error) {
	var row sessionRow
	query := `
		SELECT token, user_id, company_roles, expires_at, created_at
		FROM sessions
		WHERE token = $1 AND expires_at > NOW()
	`

	err := r.db.GetContext(ctx, &row, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	// Parse company roles from JSON
	var companyRoles auth.CompanyRoles
	if len(row.CompanyRoles) > 0 {
		if err := json.Unmarshal(row.CompanyRoles, &companyRoles); err != nil {
			return nil, err
		}
	}

	session := &auth.Session{
		Token:        row.Token,
		UserID:       row.UserID,
		CompanyRoles: companyRoles,
	}

	if row.ExpiresAt.Valid {
		session.ExpiresAt = row.ExpiresAt.Time
	}
	if row.CreatedAt.Valid {
		session.CreatedAt = row.CreatedAt.Time
	}

	return session, nil
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

// GetUserCompanies retrieves all companies a user has access to
func (r *repository) GetUserCompanies(ctx context.Context, userID int64) ([]auth.UserCompany, error) {
	var companies []auth.UserCompany
	query := `
		SELECT uc.company_id, mc.name as company_name, uc.role
		FROM user_companies uc
		INNER JOIN mining_companies mc ON mc.id = uc.company_id
		WHERE uc.user_id = $1 AND mc.active = true
		ORDER BY mc.name
	`

	err := r.db.SelectContext(ctx, &companies, query, userID)
	if err != nil {
		return nil, err
	}

	return companies, nil
}

// UserHasCompanyAccess checks if a user has access to a specific company
func (r *repository) UserHasCompanyAccess(ctx context.Context, userID int64, companyID int64) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM user_companies uc
		INNER JOIN mining_companies mc ON mc.id = uc.company_id
		WHERE uc.user_id = $1 AND uc.company_id = $2 AND mc.active = true
	`

	err := r.db.GetContext(ctx, &count, query, userID, companyID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserCompanyRole returns the user's role in a specific company
// Returns empty string if user has no access to the company
func (r *repository) GetUserCompanyRole(ctx context.Context, userID int64, companyID int64) (string, error) {
	var role string
	query := `
		SELECT uc.role
		FROM user_companies uc
		INNER JOIN mining_companies mc ON mc.id = uc.company_id
		WHERE uc.user_id = $1 AND uc.company_id = $2 AND mc.active = true
	`

	err := r.db.GetContext(ctx, &role, query, userID, companyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil // No access
		}
		return "", err
	}

	return role, nil
}

// UpdateUser updates an existing user
func (r *repository) UpdateUser(ctx context.Context, user *auth.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, work_area = $3, updated_at = NOW()
		WHERE id = $4 AND active = true
	`

	result, err := r.db.ExecContext(ctx, query, user.FirstName, user.LastName, user.WorkArea, user.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateUserPassword updates only the user's password hash
func (r *repository) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2 AND active = true
	`

	result, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeactivateUser soft deletes a user by setting active = false
func (r *repository) DeactivateUser(ctx context.Context, userID int64) error {
	query := `UPDATE users SET active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsersByCompany retrieves paginated list of users for a specific company
func (r *repository) ListUsersByCompany(ctx context.Context, companyID int64, page, size int) ([]*auth.User, int, error) {
	offset := (page - 1) * size

	var users []*auth.User
	query := `
		SELECT u.id, u.first_name, u.last_name, u.dni, u.birth_date, u.work_area, 
		       u.password_hash, u.active, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_companies uc ON u.id = uc.user_id
		WHERE uc.company_id = $1 AND u.active = true
		ORDER BY u.id
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &users, query, companyID, size, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM users u
		INNER JOIN user_companies uc ON u.id = uc.user_id
		WHERE uc.company_id = $1 AND u.active = true
	`
	err = r.db.GetContext(ctx, &total, countQuery, companyID)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// AssignUserToCompany assigns a user to a company with a specific role
func (r *repository) AssignUserToCompany(ctx context.Context, userID int64, companyID int64, role string) error {
	query := `
		INSERT INTO user_companies (user_id, company_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, company_id) DO UPDATE SET role = EXCLUDED.role
	`

	_, err := r.db.ExecContext(ctx, query, userID, companyID, role)
	return err
}

// UpdateUserCompanyRole updates the user's role in a company
func (r *repository) UpdateUserCompanyRole(ctx context.Context, userID int64, companyID int64, role string) error {
	query := `
		UPDATE user_companies
		SET role = $1
		WHERE user_id = $2 AND company_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, role, userID, companyID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user is not assigned to this company")
	}

	return nil
}

// RemoveUserFromCompany removes a user's access to a company
func (r *repository) RemoveUserFromCompany(ctx context.Context, userID int64, companyID int64) error {
	query := `DELETE FROM user_companies WHERE user_id = $1 AND company_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, companyID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user is not assigned to this company")
	}

	return nil
}
