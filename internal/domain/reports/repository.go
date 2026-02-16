package reports

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/gmhafiz/go8/internal/domain/data"
)

var (
	ErrCompanyNotFound = errors.New("company not found")
	ErrReportNotFound  = errors.New("report not found")
)

type Repository interface {
	GetCompanyName(ctx context.Context, companyID int64) (string, error)
	GetCompanyConfig(ctx context.Context, companyID int64) (*CompanyConfig, error)
	GetPBRData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.PBRData, error)
	GetDoreData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.DoreData, error)
	GetOPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.OPEXData, error)
	GetCAPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.CAPEXData, error)
	GetFinancialData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.FinancialData, error)
	GetProductionData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.ProductionData, error)
	GetRevenueData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.RevenueData, error)
	GetMineralMap(ctx context.Context) (map[int]struct{ Code, Name string }, error) // mineral_id -> {code, name}

	// Saved reports (for scenario comparison)
	SaveReport(ctx context.Context, report *SavedReport) error
	ListSavedReports(ctx context.Context, companyID int64, year int) ([]*SavedReport, error)
	GetSavedReportsByIDs(ctx context.Context, ids []int64) ([]*SavedReport, error)
	GetReportCompanyID(ctx context.Context, reportID int64) (int64, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetCompanyName(ctx context.Context, companyID int64) (string, error) {
	var name string
	query := `SELECT name FROM mining_companies WHERE id = $1`

	err := r.db.GetContext(ctx, &name, query, companyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrCompanyNotFound
		}
		return "", err
	}

	return name, nil
}

func (r *repository) GetCompanyConfig(ctx context.Context, companyID int64) (*CompanyConfig, error) {
	config := &CompanyConfig{
		MiningType: "both",    // Default
		Minerals:   []string{}, // Empty list by default
	}

	// Get mining type from company_settings
	var miningType sql.NullString
	settingsQuery := `SELECT mining_type FROM company_settings WHERE company_id = $1`
	err := r.db.GetContext(ctx, &miningType, settingsQuery, companyID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if miningType.Valid && miningType.String != "" {
		config.MiningType = miningType.String
	}

	// Get minerals assigned to company
	var mineralCodes []string
	mineralsQuery := `
		SELECT m.code
		FROM minerals m
		INNER JOIN company_minerals cm ON m.id = cm.mineral_id
		WHERE cm.company_id = $1 AND m.active = true
		ORDER BY m.code
	`
	err = r.db.SelectContext(ctx, &mineralCodes, mineralsQuery, companyID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if len(mineralCodes) > 0 {
		config.Minerals = mineralCodes
	}

	return config, nil
}

func (r *repository) GetPBRData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.PBRData, error) {
	var records []*data.PBRData

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	query := `
		SELECT id, company_id, date,
		       open_pit_ore_t, underground_ore_t, ore_mined_t,
		       waste_mined_t, stripping_ratio,
		       mining_grade_silver_gpt, mining_grade_gold_gpt,
		       open_pit_grade_silver_gpt, underground_grade_silver_gpt,
		       open_pit_grade_gold_gpt, underground_grade_gold_gpt,
		       primary_development_m, secondary_development_opex_m, expansionary_development_m, developments_m,
		       total_tonnes_processed, feed_grade_silver_gpt, feed_grade_gold_gpt,
		       recovery_rate_silver_pct, recovery_rate_gold_pct,
		       full_time_employees, contractors, total_headcount,
		       data_type, version, created_by, created_at
		FROM pbr_data
		WHERE company_id = $1 AND date >= $2 AND date <= $3 AND data_type = $4 AND version = $5 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, startDate, endDate, dataType, version)
	return records, err
}

func (r *repository) GetDoreData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.DoreData, error) {
	var records []*data.DoreData

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	query := `
		SELECT id, company_id, date, dore_produced_oz, silver_grade_pct, gold_grade_pct,
		       pbr_price_silver, pbr_price_gold, realized_price_silver, realized_price_gold,
		       silver_adjustment_oz, gold_adjustment_oz, ag_deductions_pct, au_deductions_pct,
		       treatment_charge, refining_deductions_au, streaming,
		       data_type, version, created_by, created_at
		FROM dore_data
		WHERE company_id = $1 AND date >= $2 AND date <= $3 AND data_type = $4 AND version = $5 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, startDate, endDate, dataType, version)
	return records, err
}

func (r *repository) GetOPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.OPEXData, error) {
	var records []*data.OPEXData

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	query := `
		SELECT id, company_id, date, cost_center, subcategory, expense_type,
		       amount, currency, data_type, version, created_by, created_at
		FROM opex_data
		WHERE company_id = $1 AND date >= $2 AND date <= $3 AND data_type = $4 AND version = $5 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, startDate, endDate, dataType, version)
	return records, err
}

func (r *repository) GetCAPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.CAPEXData, error) {
	var records []*data.CAPEXData

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	query := `
		SELECT id, company_id, date, category, car_number, project_name, type,
		       amount, accretion_of_mine_closure_liability, currency, data_type, version, created_by, created_at
		FROM capex_data
		WHERE company_id = $1 AND date >= $2 AND date <= $3 AND data_type = $4 AND version = $5 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, startDate, endDate, dataType, version)
	return records, err
}

func (r *repository) GetFinancialData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.FinancialData, error) {
	var records []*data.FinancialData

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	query := `
		SELECT id, company_id, date, shipping_selling, sales_taxes, royalties,
		       other_sales_deductions, other_adjustments, currency, data_type, version, created_by, created_at
		FROM financial_data
		WHERE company_id = $1 AND date >= $2 AND date <= $3 AND data_type = $4 AND version = $5 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, startDate, endDate, dataType, version)
	return records, err
}

func (r *repository) GetProductionData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.ProductionData, error) {
	var records []*data.ProductionData

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	query := `
		SELECT id, company_id, date, mineral_id, quantity, unit,
		       data_type, version, description, created_by, created_at
		FROM production_data
		WHERE company_id = $1 AND date >= $2 AND date <= $3 AND data_type = $4 AND version = $5 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, startDate, endDate, dataType, version)
	return records, err
}

func (r *repository) GetRevenueData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.RevenueData, error) {
	var records []*data.RevenueData

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	query := `
		SELECT id, company_id, date, mineral_id, quantity_sold, unit_price,
		       currency, data_type, version, description, created_by, created_at
		FROM revenue_data
		WHERE company_id = $1 AND date >= $2 AND date <= $3 AND data_type = $4 AND version = $5 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, startDate, endDate, dataType, version)
	return records, err
}

// SaveReport saves a report snapshot
func (r *repository) SaveReport(ctx context.Context, report *SavedReport) error {
	// Marshal ReportData to JSON
	reportJSON, err := json.Marshal(report.ReportData)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO saved_reports (company_id, name, description, year, 
		                           budget_version, report_data, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	return r.db.QueryRowContext(ctx, query,
		report.CompanyID,
		report.Name,
		report.Description,
		report.Year,
		report.BudgetVersion,
		reportJSON,
		report.CreatedBy,
	).Scan(&report.ID, &report.CreatedAt)
}

// ListSavedReports retrieves saved reports for a company and year
func (r *repository) ListSavedReports(ctx context.Context, companyID int64, year int) ([]*SavedReport, error) {
	var reports []*SavedReport

	query := `
		SELECT id, company_id, name, description, year,
		       budget_version, report_data, created_by, created_at
		FROM saved_reports
		WHERE company_id = $1 AND year = $2
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &reports, query, companyID, year)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB to struct
	for _, report := range reports {
		if err := json.Unmarshal(report.ReportDataRaw, &report.ReportData); err != nil {
			return nil, err
		}
	}

	return reports, nil
}

// GetSavedReportsByIDs retrieves multiple saved reports by IDs
func (r *repository) GetSavedReportsByIDs(ctx context.Context, ids []int64) ([]*SavedReport, error) {
	var reports []*SavedReport

	query := `
		SELECT id, company_id, name, description, year,
		       budget_version, report_data, created_by, created_at
		FROM saved_reports
		WHERE id = ANY($1)
		ORDER BY id
	`

	err := r.db.SelectContext(ctx, &reports, query, ids)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB to struct
	for _, report := range reports {
		if err := json.Unmarshal(report.ReportDataRaw, &report.ReportData); err != nil {
			return nil, err
		}
	}

	return reports, nil
}

// GetMineralMap returns a map of mineral_id -> {code, name}
func (r *repository) GetMineralMap(ctx context.Context) (map[int]struct{ Code, Name string }, error) {
	type mineralRow struct {
		ID   int    `db:"id"`
		Code string `db:"code"`
		Name string `db:"name"`
	}

	var minerals []mineralRow
	query := `SELECT id, code, name FROM minerals WHERE active = true`
	err := r.db.SelectContext(ctx, &minerals, query)
	if err != nil {
		return nil, err
	}

	mineralMap := make(map[int]struct{ Code, Name string })
	for _, m := range minerals {
		mineralMap[m.ID] = struct{ Code, Name string }{Code: m.Code, Name: m.Name}
	}

	return mineralMap, nil
}

// GetReportCompanyID returns the company ID for a saved report
func (r *repository) GetReportCompanyID(ctx context.Context, reportID int64) (int64, error) {
	var companyID int64
	query := `SELECT company_id FROM saved_reports WHERE id = $1`

	err := r.db.GetContext(ctx, &companyID, query, reportID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrReportNotFound
		}
		return 0, err
	}

	return companyID, nil
}
