package config

// CreateCompanyRequest represents request to create a company
type CreateCompanyRequest struct {
	Name         string `json:"name" validate:"required"`
	LegalName    string `json:"legal_name" validate:"required"`
	TaxID        string `json:"tax_id" validate:"required"`
	Address      string `json:"address"`
	ContactEmail string `json:"contact_email" validate:"omitempty,email"`
	ContactPhone string `json:"contact_phone"`
	// Settings (optional on creation)
	MiningType        string   `json:"mining_type" validate:"omitempty,oneof=open_pit underground both"`
	Country           string   `json:"country"`
	RoyaltyPercentage *float64 `json:"royalty_percentage" validate:"omitempty,gte=0,lte=100"`
}

// UpdateCompanyRequest represents request to update a company
type UpdateCompanyRequest struct {
	Name         string `json:"name"`
	LegalName    string `json:"legal_name"`
	Address      string `json:"address"`
	ContactEmail string `json:"contact_email" validate:"omitempty,email"`
	ContactPhone string `json:"contact_phone"`
	Active       *bool  `json:"active"`
}

// UpdateCompanySettingsRequest represents request to update company settings
type UpdateCompanySettingsRequest struct {
	MiningType        string   `json:"mining_type" validate:"omitempty,oneof=open_pit underground both"`
	Country           string   `json:"country"`
	RoyaltyPercentage *float64 `json:"royalty_percentage" validate:"omitempty,gte=0,lte=100"`
	Notes             string   `json:"notes"`
}

// AssignMineralsRequest represents request to assign minerals to a company
type AssignMineralsRequest struct {
	MineralIDs []int `json:"mineral_ids" validate:"required,min=1"`
}

// CreateMineralRequest represents request to create a mineral
type CreateMineralRequest struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required,min=1,max=10"`
	Description string `json:"description"`
}

// UpdateMineralRequest represents request to update a mineral
type UpdateMineralRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      *bool  `json:"active"`
}
