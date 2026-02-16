package reports

import (
	"encoding/json"
	"time"
)

// SavedReport represents a saved report snapshot for comparison
type SavedReport struct {
	ID            int64           `db:"id" json:"id"`
	CompanyID     int64           `db:"company_id" json:"company_id"`
	Name          string          `db:"name" json:"name"`
	Description   string          `db:"description" json:"description,omitempty"`
	Year          int             `db:"year" json:"year"`
	BudgetVersion int             `db:"budget_version" json:"budget_version"`
	ReportData    SummaryReport   `db:"-" json:"report_data"` // Handled separately
	ReportDataRaw json.RawMessage `db:"report_data" json:"-"`
	CreatedBy     int64           `db:"created_by" json:"created_by"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
}

// SaveReportRequest represents request to save a report (always saves full year)
type SaveReportRequest struct {
	Name          string `json:"name" validate:"required,min=3,max=200"`
	Description   string `json:"description"`
	CompanyID     int64  `json:"company_id" validate:"required,gt=0"`
	Year          int    `json:"year" validate:"required,gte=2000"`
	BudgetVersion int    `json:"budget_version" validate:"required,gte=1"` // Required: budget version to use
}

// CompareReportsRequest represents request to compare multiple saved reports
type CompareReportsRequest struct {
	ReportIDs []int64 `json:"report_ids" validate:"required,min=2,max=5"`
}

// CompareReportsResponse represents comparison of multiple reports
type CompareReportsResponse struct {
	Reports    []SavedReport `json:"reports"`
	Comparison interface{}   `json:"comparison"` // Side-by-side comparison data
}
