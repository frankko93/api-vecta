package data

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	// Production
	InsertProductionBulk(ctx context.Context, records []*ProductionData) error
	ListPBRData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*PBRData, error)
	SoftDeletePBRData(ctx context.Context, id int64) error

	// Dore
	InsertDoreBulk(ctx context.Context, records []*DoreData) error
	ListDoreData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*DoreData, error)
	SoftDeleteDoreData(ctx context.Context, id int64) error

	// PBR
	InsertPBRBulk(ctx context.Context, records []*PBRData) error
	GetPBRByDate(ctx context.Context, companyID int64, date time.Time, dataType string, version int) (*PBRData, error)

	// OPEX
	InsertOPEXBulk(ctx context.Context, records []*OPEXData) error
	ListOPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*OPEXData, error)
	SoftDeleteOPEXData(ctx context.Context, id int64) error

	// CAPEX
	InsertCAPEXBulk(ctx context.Context, records []*CAPEXData) error
	ListCAPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*CAPEXData, error)
	SoftDeleteCAPEXData(ctx context.Context, id int64) error

	// Revenue
	InsertRevenueBulk(ctx context.Context, records []*RevenueData) error

	// Financial
	InsertFinancialBulk(ctx context.Context, records []*FinancialData) error
	ListFinancialData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*FinancialData, error)
	SoftDeleteFinancialData(ctx context.Context, id int64) error

	// Helpers
	GetMineralCodeMap(ctx context.Context) (map[string]int, error)
	CompanyExists(ctx context.Context, companyID int64) (bool, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) InsertProductionBulk(ctx context.Context, records []*ProductionData) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO production_data (company_id, date, mineral_id, quantity, unit, data_type, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, query,
			record.CompanyID,
			record.Date,
			record.MineralID,
			record.Quantity,
			record.Unit,
			record.DataType,
			record.CreatedBy,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) InsertDoreBulk(ctx context.Context, records []*DoreData) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO dore_data (
			company_id, date, dore_produced_oz, silver_grade_pct, gold_grade_pct,
			pbr_price_silver, pbr_price_gold, realized_price_silver, realized_price_gold,
			silver_adjustment_oz, gold_adjustment_oz, ag_deductions_pct, au_deductions_pct,
			treatment_charge, refining_deductions_au, streaming, data_type, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, query,
			record.CompanyID, record.Date, record.DoreProducedOz, record.SilverGradePct, record.GoldGradePct,
			record.PBRPriceSilver, record.PBRPriceGold, record.RealizedPriceSilver, record.RealizedPriceGold,
			record.SilverAdjustmentOz, record.GoldAdjustmentOz, record.AgDeductionsPct, record.AuDeductionsPct,
			record.TreatmentCharge, record.RefiningDeductionsAu, record.Streaming, record.DataType, record.CreatedBy,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) InsertPBRBulk(ctx context.Context, records []*PBRData) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO pbr_data (
			company_id, date,
			open_pit_ore_t, underground_ore_t, ore_mined_t,
			waste_mined_t, stripping_ratio,
			mining_grade_silver_gpt, mining_grade_gold_gpt,
			open_pit_grade_silver_gpt, underground_grade_silver_gpt,
			open_pit_grade_gold_gpt, underground_grade_gold_gpt,
			primary_development_m, secondary_development_opex_m, expansionary_development_m, developments_m,
			total_tonnes_processed, feed_grade_silver_gpt, feed_grade_gold_gpt,
			recovery_rate_silver_pct, recovery_rate_gold_pct,
			full_time_employees, contractors, total_headcount,
			data_type, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
	`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, query,
			record.CompanyID, record.Date,
			record.OpenPitOreT, record.UndergroundOreT, record.OreMinedT,
			record.WasteMinedT, record.StrippingRatio,
			record.MiningGradeSilverGpt, record.MiningGradeGoldGpt,
			record.OpenPitGradeSilverGpt, record.UndergroundGradeSilverGpt,
			record.OpenPitGradeGoldGpt, record.UndergroundGradeGoldGpt,
			record.PrimaryDevelopmentM, record.SecondaryDevelopmentOpexM, record.ExpansionaryDevelopmentM, record.DevelopmentsM,
			record.TotalTonnesProcessed, record.FeedGradeSilverGpt, record.FeedGradeGoldGpt,
			record.RecoveryRateSilverPct, record.RecoveryRateGoldPct,
			record.FullTimeEmployees, record.Contractors, record.TotalHeadcount,
			record.DataType, record.CreatedBy,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) InsertOPEXBulk(ctx context.Context, records []*OPEXData) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO opex_data (company_id, date, cost_center, subcategory, expense_type, amount, currency, data_type, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, query,
			record.CompanyID, record.Date, record.CostCenter, record.Subcategory,
			record.ExpenseType, record.Amount, record.Currency, record.DataType, record.CreatedBy,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) InsertCAPEXBulk(ctx context.Context, records []*CAPEXData) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO capex_data (company_id, date, category, car_number, project_name, type, amount, accretion_of_mine_closure_liability, currency, data_type, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, query,
			record.CompanyID, record.Date, record.Category, record.CARNumber,
			record.ProjectName, record.Type, record.Amount, record.AccretionOfMineClosureLiability, record.Currency, record.DataType, record.CreatedBy,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) InsertRevenueBulk(ctx context.Context, records []*RevenueData) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO revenue_data (company_id, date, mineral_id, quantity_sold, unit_price, currency, data_type, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, query,
			record.CompanyID, record.Date, record.MineralID,
			record.QuantitySold, record.UnitPrice, record.Currency, record.DataType, record.CreatedBy,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) InsertFinancialBulk(ctx context.Context, records []*FinancialData) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO financial_data (company_id, date, shipping_selling, sales_taxes_royalties, other_adjustments, currency, data_type, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, query,
			record.CompanyID, record.Date, record.ShippingSelling,
			record.SalesTaxesRoyalties, record.OtherAdjustments,
			record.Currency, record.DataType, record.CreatedBy,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) GetMineralCodeMap(ctx context.Context) (map[string]int, error) {
	var minerals []struct {
		ID   int    `db:"id"`
		Code string `db:"code"`
	}

	query := `SELECT id, code FROM minerals WHERE active = true`
	err := r.db.SelectContext(ctx, &minerals, query)
	if err != nil {
		return nil, err
	}

	mineralMap := make(map[string]int)
	for _, m := range minerals {
		mineralMap[m.Code] = m.ID
	}

	return mineralMap, nil
}

func (r *repository) CompanyExists(ctx context.Context, companyID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM mining_companies WHERE id = $1 AND active = true)`
	err := r.db.GetContext(ctx, &exists, query, companyID)
	return exists, err
}

// List PBR Data
func (r *repository) ListPBRData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*PBRData, error) {
	var records []*PBRData

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
		       data_type, version, description, created_by, created_at
		FROM pbr_data
		WHERE company_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND data_type = $3 
		      AND version = $4 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, year, dataType, version)
	return records, err
}

func (r *repository) SoftDeletePBRData(ctx context.Context, id int64) error {
	query := `UPDATE pbr_data SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}

	return nil
}

// GetPBRByDate gets PBR data for a specific date
func (r *repository) GetPBRByDate(ctx context.Context, companyID int64, date time.Time, dataType string, version int) (*PBRData, error) {
	var record PBRData

	// Get the year from the date
	year := date.Year()

	query := `
		SELECT id, company_id, date, ore_mined_t, waste_mined_t, developments_m,
		       total_tonnes_processed, feed_grade_silver_gpt, feed_grade_gold_gpt,
		       recovery_rate_silver_pct, recovery_rate_gold_pct, data_type, version,
		       description, created_by, created_at
		FROM pbr_data
		WHERE company_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND data_type = $3 
		      AND version = $4 AND deleted_at IS NULL
		      AND date = $5
		ORDER BY date
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &record, query, companyID, year, dataType, version, date)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// List Dore Data
func (r *repository) ListDoreData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*DoreData, error) {
	var records []*DoreData

	query := `
		SELECT id, company_id, date, dore_produced_oz, silver_grade_pct, gold_grade_pct,
		       pbr_price_silver, pbr_price_gold, realized_price_silver, realized_price_gold,
		       silver_adjustment_oz, gold_adjustment_oz, ag_deductions_pct, au_deductions_pct,
		       treatment_charge, refining_deductions_au, streaming, data_type, version,
		       description, created_by, created_at
		FROM dore_data
		WHERE company_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND data_type = $3 
		      AND version = $4 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, year, dataType, version)
	return records, err
}

func (r *repository) SoftDeleteDoreData(ctx context.Context, id int64) error {
	query := `UPDATE dore_data SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}

	return nil
}

// List OPEX Data
func (r *repository) ListOPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*OPEXData, error) {
	var records []*OPEXData

	query := `
		SELECT id, company_id, date, cost_center, subcategory, expense_type,
		       amount, currency, data_type, version, description, created_by, created_at
		FROM opex_data
		WHERE company_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND data_type = $3 
		      AND version = $4 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, year, dataType, version)
	return records, err
}

func (r *repository) SoftDeleteOPEXData(ctx context.Context, id int64) error {
	query := `UPDATE opex_data SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}

	return nil
}

// List CAPEX Data
func (r *repository) ListCAPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*CAPEXData, error) {
	var records []*CAPEXData

	query := `
		SELECT id, company_id, date, category, car_number, project_name, type,
		       amount, accretion_of_mine_closure_liability, currency, data_type, version, description, created_by, created_at
		FROM capex_data
		WHERE company_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND data_type = $3 
		      AND version = $4 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, year, dataType, version)
	return records, err
}

func (r *repository) SoftDeleteCAPEXData(ctx context.Context, id int64) error {
	query := `UPDATE capex_data SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}

	return nil
}

// List Financial Data
func (r *repository) ListFinancialData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*FinancialData, error) {
	var records []*FinancialData

	query := `
		SELECT id, company_id, date, shipping_selling, sales_taxes_royalties,
		       other_adjustments, currency, data_type, version, description, created_by, created_at
		FROM financial_data
		WHERE company_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND data_type = $3 
		      AND version = $4 AND deleted_at IS NULL
		ORDER BY date
	`

	err := r.db.SelectContext(ctx, &records, query, companyID, year, dataType, version)
	return records, err
}

func (r *repository) SoftDeleteFinancialData(ctx context.Context, id int64) error {
	query := `UPDATE financial_data SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}

	return nil
}
