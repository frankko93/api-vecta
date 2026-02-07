package minerals

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/gmhafiz/go8/internal/domain/config"
)

var (
	ErrMineralNotFound = errors.New("mineral not found")
	ErrCodeExists      = errors.New("mineral code already exists")
)

type Repository interface {
	List(ctx context.Context) ([]*config.Mineral, error)
	GetByID(ctx context.Context, id int) (*config.Mineral, error)
	Create(ctx context.Context, mineral *config.Mineral) error
	Update(ctx context.Context, mineral *config.Mineral) error
	Delete(ctx context.Context, id int) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]*config.Mineral, error) {
	var minerals []*config.Mineral
	query := `
		SELECT id, name, code, description, active, created_at, updated_at
		FROM minerals
		WHERE active = true
		ORDER BY name
	`
	
	err := r.db.SelectContext(ctx, &minerals, query)
	return minerals, err
}

func (r *repository) GetByID(ctx context.Context, id int) (*config.Mineral, error) {
	var mineral config.Mineral
	query := `
		SELECT id, name, code, description, active, created_at, updated_at
		FROM minerals
		WHERE id = $1
	`
	
	err := r.db.GetContext(ctx, &mineral, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMineralNotFound
		}
		return nil, err
	}
	
	return &mineral, nil
}

func (r *repository) Create(ctx context.Context, mineral *config.Mineral) error {
	query := `
		INSERT INTO minerals (name, code, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(
		ctx,
		query,
		mineral.Name,
		mineral.Code,
		mineral.Description,
	).Scan(&mineral.ID, &mineral.CreatedAt, &mineral.UpdatedAt)
	
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "minerals_code_key"` {
			return ErrCodeExists
		}
		return err
	}
	
	mineral.Active = true
	return nil
}

func (r *repository) Update(ctx context.Context, mineral *config.Mineral) error {
	query := `
		UPDATE minerals
		SET name = $2, description = $3, active = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(
		ctx,
		query,
		mineral.ID,
		mineral.Name,
		mineral.Description,
		mineral.Active,
	).Scan(&mineral.UpdatedAt)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrMineralNotFound
		}
		return err
	}
	
	return nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	query := `UPDATE minerals SET active = false WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return ErrMineralNotFound
	}
	
	return nil
}

