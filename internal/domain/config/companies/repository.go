package companies

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/gmhafiz/go8/internal/domain/config"
)

var (
	ErrCompanyNotFound = errors.New("company not found")
	ErrTaxIDExists     = errors.New("tax ID already exists")
)

type Repository interface {
	// Companies
	List(ctx context.Context) ([]*config.MiningCompany, error)
	GetByID(ctx context.Context, id int64) (*config.MiningCompany, error)
	Create(ctx context.Context, company *config.MiningCompany) error
	Update(ctx context.Context, company *config.MiningCompany) error
	Delete(ctx context.Context, id int64) error

	// Minerals assignment
	GetCompanyMinerals(ctx context.Context, companyID int64) ([]*config.Mineral, error)
	AssignMinerals(ctx context.Context, companyID int64, mineralIDs []int) error

	// Settings
	GetSettings(ctx context.Context, companyID int64) (*config.CompanySettings, error)
	UpsertSettings(ctx context.Context, settings *config.CompanySettings) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]*config.MiningCompany, error) {
	var companies []*config.MiningCompany
	query := `
		SELECT id, name, legal_name, tax_id, address, contact_email, contact_phone, 
		       active, created_at, updated_at
		FROM mining_companies
		WHERE active = true
		ORDER BY name
	`

	err := r.db.SelectContext(ctx, &companies, query)
	return companies, err
}

func (r *repository) GetByID(ctx context.Context, id int64) (*config.MiningCompany, error) {
	var company config.MiningCompany
	query := `
		SELECT id, name, legal_name, tax_id, address, contact_email, contact_phone,
		       active, created_at, updated_at
		FROM mining_companies
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &company, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}

	return &company, nil
}

func (r *repository) Create(ctx context.Context, company *config.MiningCompany) error {
	query := `
		INSERT INTO mining_companies (name, legal_name, tax_id, address, contact_email, contact_phone)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		company.Name,
		company.LegalName,
		company.TaxID,
		company.Address,
		company.ContactEmail,
		company.ContactPhone,
	).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "mining_companies_tax_id_key"` {
			return ErrTaxIDExists
		}
		return err
	}

	company.Active = true
	return nil
}

func (r *repository) Update(ctx context.Context, company *config.MiningCompany) error {
	query := `
		UPDATE mining_companies
		SET name = $2, legal_name = $3, address = $4, contact_email = $5, 
		    contact_phone = $6, active = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		company.ID,
		company.Name,
		company.LegalName,
		company.Address,
		company.ContactEmail,
		company.ContactPhone,
		company.Active,
	).Scan(&company.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCompanyNotFound
		}
		return err
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE mining_companies SET active = false WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrCompanyNotFound
	}

	return nil
}

func (r *repository) GetCompanyMinerals(ctx context.Context, companyID int64) ([]*config.Mineral, error) {
	var minerals []*config.Mineral
	query := `
		SELECT m.id, m.name, m.code, m.description, m.active, m.created_at, m.updated_at
		FROM minerals m
		INNER JOIN company_minerals cm ON m.id = cm.mineral_id
		WHERE cm.company_id = $1 AND m.active = true
		ORDER BY m.name
	`

	err := r.db.SelectContext(ctx, &minerals, query, companyID)
	return minerals, err
}

func (r *repository) AssignMinerals(ctx context.Context, companyID int64, mineralIDs []int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing assignments
	_, err = tx.ExecContext(ctx, `DELETE FROM company_minerals WHERE company_id = $1`, companyID)
	if err != nil {
		return err
	}

	// Insert new assignments
	insertQuery := `INSERT INTO company_minerals (company_id, mineral_id) VALUES ($1, $2)`
	for _, mineralID := range mineralIDs {
		_, err = tx.ExecContext(ctx, insertQuery, companyID, mineralID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) GetSettings(ctx context.Context, companyID int64) (*config.CompanySettings, error) {
	var settings config.CompanySettings
	query := `
		SELECT company_id, mining_type, country, royalty_percentage, notes, created_at, updated_at
		FROM company_settings
		WHERE company_id = $1
	`

	err := r.db.GetContext(ctx, &settings, query, companyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No settings yet, not an error
		}
		return nil, err
	}

	return &settings, nil
}

func (r *repository) UpsertSettings(ctx context.Context, settings *config.CompanySettings) error {
	query := `
		INSERT INTO company_settings (company_id, mining_type, country, royalty_percentage, notes)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (company_id) DO UPDATE
		SET mining_type = $2, country = $3, royalty_percentage = $4, notes = $5, updated_at = CURRENT_TIMESTAMP
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		settings.CompanyID,
		settings.MiningType,
		settings.Country,
		settings.RoyaltyPercentage,
		settings.Notes,
	).Scan(&settings.CreatedAt, &settings.UpdatedAt)

	return err
}
