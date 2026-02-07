package config

import "time"

// MiningCompany represents a mining company
type MiningCompany struct {
	ID           int64     `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	LegalName    string    `db:"legal_name" json:"legal_name"`
	TaxID        string    `db:"tax_id" json:"tax_id"`
	Address      string    `db:"address" json:"address"`
	ContactEmail string    `db:"contact_email" json:"contact_email"`
	ContactPhone string    `db:"contact_phone" json:"contact_phone"`
	Active       bool      `db:"active" json:"active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Mineral represents a mineral type
type Mineral struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Code        string    `db:"code" json:"code"`
	Description string    `db:"description" json:"description"`
	Active      bool      `db:"active" json:"active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CompanySettings represents company-specific settings
type CompanySettings struct {
	CompanyID         int64     `db:"company_id" json:"company_id"`
	MiningType        string    `db:"mining_type" json:"mining_type"` // "open_pit", "underground", "both"
	Country           string    `db:"country" json:"country"`
	RoyaltyPercentage float64   `db:"royalty_percentage" json:"royalty_percentage"`
	Notes             string    `db:"notes" json:"notes"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// CompanyWithDetails includes company info with minerals and settings
type CompanyWithDetails struct {
	MiningCompany
	Minerals []Mineral        `json:"minerals"`
	Settings *CompanySettings `json:"settings,omitempty"`
}
