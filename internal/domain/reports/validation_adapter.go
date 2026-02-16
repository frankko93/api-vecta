package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/gmhafiz/go8/internal/domain/data"
)

// reportsRepositoryAdapter adapts reports.Repository to data.Repository interface for validation
type reportsRepositoryAdapter struct {
	repo Repository
}

// Implement data.Repository interface methods needed for validation
func (a *reportsRepositoryAdapter) ListPBRData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.PBRData, error) {
	return a.repo.GetPBRData(ctx, companyID, year, dataType, version)
}

func (a *reportsRepositoryAdapter) ListDoreData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.DoreData, error) {
	return a.repo.GetDoreData(ctx, companyID, year, dataType, version)
}

func (a *reportsRepositoryAdapter) ListOPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.OPEXData, error) {
	return a.repo.GetOPEXData(ctx, companyID, year, dataType, version)
}

func (a *reportsRepositoryAdapter) ListCAPEXData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.CAPEXData, error) {
	return a.repo.GetCAPEXData(ctx, companyID, year, dataType, version)
}

func (a *reportsRepositoryAdapter) ListFinancialData(ctx context.Context, companyID int64, year int, dataType string, version int) ([]*data.FinancialData, error) {
	return a.repo.GetFinancialData(ctx, companyID, year, dataType, version)
}

// Unused methods required by data.Repository interface (not needed for validation)
func (a *reportsRepositoryAdapter) CompanyExists(ctx context.Context, companyID int64) (bool, error) {
	_, err := a.repo.GetCompanyName(ctx, companyID)
	return err == nil, err
}

func (a *reportsRepositoryAdapter) GetMineralCodeMap(ctx context.Context) (map[string]int, error) {
	// Not needed for validation, return empty map
	return make(map[string]int), nil
}

func (a *reportsRepositoryAdapter) GetPBRByDate(ctx context.Context, companyID int64, date time.Time, dataType string, version int) (*data.PBRData, error) {
	// Not needed for validation, but required by interface
	// Could be implemented if needed
	return nil, fmt.Errorf("not implemented")
}

func (a *reportsRepositoryAdapter) InsertProductionBulk(ctx context.Context, records []*data.ProductionData) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) InsertDoreBulk(ctx context.Context, records []*data.DoreData) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) InsertPBRBulk(ctx context.Context, records []*data.PBRData) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) InsertOPEXBulk(ctx context.Context, records []*data.OPEXData) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) InsertCAPEXBulk(ctx context.Context, records []*data.CAPEXData) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) InsertRevenueBulk(ctx context.Context, records []*data.RevenueData) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) InsertFinancialBulk(ctx context.Context, records []*data.FinancialData) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) SoftDeletePBRData(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) SoftDeleteDoreData(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) SoftDeleteOPEXData(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) SoftDeleteCAPEXData(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

func (a *reportsRepositoryAdapter) SoftDeleteFinancialData(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented - read-only adapter")
}

// validateCrossFile performs mandatory cross-file validations
func validateCrossFile(ctx context.Context, repo Repository, companyID int64, year int) error {
	adapter := &reportsRepositoryAdapter{repo: repo}

	// Validate month alignment for actual data
	if err := data.ValidateMonthAlignment(ctx, adapter, companyID, year, "actual", 1); err != nil {
		return err
	}

	// Validate month alignment for budget data
	if err := data.ValidateMonthAlignment(ctx, adapter, companyID, year, "budget", 1); err != nil {
		return err
	}

	// Validate Dore dependencies for actual data
	if err := data.ValidateDoreDependencies(ctx, adapter, companyID, year, "actual", 1); err != nil {
		return err
	}

	// Validate Dore dependencies for budget data
	if err := data.ValidateDoreDependencies(ctx, adapter, companyID, year, "budget", 1); err != nil {
		return err
	}

	return nil
}
