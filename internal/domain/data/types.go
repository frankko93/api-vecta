package data

import "errors"

// DataImportType represents the type of data being imported
type DataImportType string

const (
	ImportProduction DataImportType = "production"
	ImportDore       DataImportType = "dore"
	ImportPBR        DataImportType = "pbr"
	ImportOPEX       DataImportType = "opex"
	ImportCAPEX      DataImportType = "capex"
	ImportRevenue    DataImportType = "revenue"
	ImportFinancial  DataImportType = "financial"
)

// DataType represents if data is actual or budget
type DataType string

const (
	DataTypeActual DataType = "actual"
	DataTypeBudget DataType = "budget"
)

// IsValid validates data type
func (dt DataType) IsValid() bool {
	switch dt {
	case DataTypeActual, DataTypeBudget:
		return true
	}
	return false
}

// IsValid validates if the import type is supported
func (t DataImportType) IsValid() bool {
	switch t {
	case ImportProduction, ImportDore, ImportPBR, ImportOPEX, ImportCAPEX, ImportRevenue, ImportFinancial:
		return true
	}
	return false
}

// Currency represents supported currencies
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyARS Currency = "ARS"
)

// IsValid validates if currency is supported
func (c Currency) IsValid() bool {
	switch c {
	case CurrencyUSD, CurrencyARS:
		return true
	}
	return false
}

// CostCenter represents OPEX cost centers
type CostCenter string

const (
	CostCenterMine       CostCenter = "Mine"
	CostCenterProcessing CostCenter = "Processing"
	CostCenterGA         CostCenter = "G&A"
	CostCenterTransport  CostCenter = "Transport & Shipping"
)

// IsValid validates cost center
func (cc CostCenter) IsValid() bool {
	switch cc {
	case CostCenterMine, CostCenterProcessing, CostCenterGA, CostCenterTransport:
		return true
	}
	return false
}

// ExpenseType represents OPEX expense types
type ExpenseType string

const (
	ExpenseLabour     ExpenseType = "Labour"
	ExpenseMaterials  ExpenseType = "Materials"
	ExpenseThirdParty ExpenseType = "Third Party"
	ExpenseOther      ExpenseType = "Other"
)

// IsValid validates expense type
func (et ExpenseType) IsValid() bool {
	switch et {
	case ExpenseLabour, ExpenseMaterials, ExpenseThirdParty, ExpenseOther:
		return true
	}
	return false
}

// CapexType represents CAPEX project types
type CapexType string

const (
	CapexSustaining CapexType = "sustaining"
	CapexProject    CapexType = "project"
	CapexLeasing    CapexType = "leasing"
	CapexAccretion  CapexType = "accretion"
)

// IsValid validates capex type
func (ct CapexType) IsValid() bool {
	switch ct {
	case CapexSustaining, CapexProject, CapexLeasing, CapexAccretion:
		return true
	}
	return false
}

// ValidationError represents a validation error for a specific row
type ValidationError struct {
	Row    int    `json:"row"`
	Column string `json:"column,omitempty"`
	Error  string `json:"error"`
}

var (
	ErrInvalidDataType  = errors.New("invalid data type")
	ErrInvalidCSVFormat = errors.New("invalid CSV format")
	ErrMissingHeaders   = errors.New("missing required headers")
	ErrCompanyNotFound  = errors.New("company not found")
	ErrMineralNotFound  = errors.New("mineral not found")
	ErrValidationFailed = errors.New("validation failed")
)
