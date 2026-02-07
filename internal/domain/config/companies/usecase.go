package companies

import (
	"context"

	"github.com/gmhafiz/go8/internal/domain/config"
)

type UseCase interface {
	List(ctx context.Context) ([]*config.MiningCompany, error)
	GetByID(ctx context.Context, id int64) (*config.CompanyWithDetails, error)
	Create(ctx context.Context, req *config.CreateCompanyRequest) (*config.MiningCompany, error)
	Update(ctx context.Context, id int64, req *config.UpdateCompanyRequest) (*config.MiningCompany, error)
	Delete(ctx context.Context, id int64) error
	AssignMinerals(ctx context.Context, companyID int64, req *config.AssignMineralsRequest) error
	UpdateSettings(ctx context.Context, companyID int64, req *config.UpdateCompanySettingsRequest) (*config.CompanySettings, error)
	GetAvailableUnits(ctx context.Context) []map[string]string
}

type useCase struct {
	repo Repository
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) List(ctx context.Context) ([]*config.MiningCompany, error) {
	return uc.repo.List(ctx)
}

func (uc *useCase) GetByID(ctx context.Context, id int64) (*config.CompanyWithDetails, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	minerals, err := uc.repo.GetCompanyMinerals(ctx, id)
	if err != nil {
		return nil, err
	}

	settings, err := uc.repo.GetSettings(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert []*Mineral to []Mineral
	mineralList := make([]config.Mineral, 0)
	for _, m := range minerals {
		if m != nil {
			mineralList = append(mineralList, *m)
		}
	}

	result := &config.CompanyWithDetails{
		MiningCompany: *company,
		Minerals:      mineralList,
		Settings:      settings,
	}

	return result, nil
}

func (uc *useCase) Create(ctx context.Context, req *config.CreateCompanyRequest) (*config.MiningCompany, error) {
	company := &config.MiningCompany{
		Name:         req.Name,
		LegalName:    req.LegalName,
		TaxID:        req.TaxID,
		Address:      req.Address,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
	}

	err := uc.repo.Create(ctx, company)
	if err != nil {
		return nil, err
	}

	// Create settings if provided
	if req.MiningType != "" || req.Country != "" || req.RoyaltyPercentage != nil {
		settings := &config.CompanySettings{
			CompanyID:  company.ID,
			MiningType: req.MiningType,
			Country:    req.Country,
		}
		if req.MiningType == "" {
			settings.MiningType = "underground" // default
		}
		if req.RoyaltyPercentage != nil {
			settings.RoyaltyPercentage = *req.RoyaltyPercentage
		}

		err = uc.repo.UpsertSettings(ctx, settings)
		if err != nil {
			return nil, err
		}
	}

	return company, nil
}

func (uc *useCase) Update(ctx context.Context, id int64, req *config.UpdateCompanyRequest) (*config.MiningCompany, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != "" {
		company.Name = req.Name
	}
	if req.LegalName != "" {
		company.LegalName = req.LegalName
	}
	if req.Address != "" {
		company.Address = req.Address
	}
	if req.ContactEmail != "" {
		company.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != "" {
		company.ContactPhone = req.ContactPhone
	}
	if req.Active != nil {
		company.Active = *req.Active
	}

	err = uc.repo.Update(ctx, company)
	if err != nil {
		return nil, err
	}

	return company, nil
}

func (uc *useCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *useCase) AssignMinerals(ctx context.Context, companyID int64, req *config.AssignMineralsRequest) error {
	// Verify company exists
	_, err := uc.repo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	return uc.repo.AssignMinerals(ctx, companyID, req.MineralIDs)
}

func (uc *useCase) UpdateSettings(ctx context.Context, companyID int64, req *config.UpdateCompanySettingsRequest) (*config.CompanySettings, error) {
	// Verify company exists
	_, err := uc.repo.GetByID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Get existing settings or create new
	settings, err := uc.repo.GetSettings(ctx, companyID)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		settings = &config.CompanySettings{CompanyID: companyID}
	}

	// Update fields
	if req.MiningType != "" {
		settings.MiningType = req.MiningType
	}
	if req.Country != "" {
		settings.Country = req.Country
	}
	if req.RoyaltyPercentage != nil {
		settings.RoyaltyPercentage = *req.RoyaltyPercentage
	}
	if req.Notes != "" {
		settings.Notes = req.Notes
	}

	err = uc.repo.UpsertSettings(ctx, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (uc *useCase) GetAvailableUnits(ctx context.Context) []map[string]string {
	return config.GetAvailableUnits()
}
