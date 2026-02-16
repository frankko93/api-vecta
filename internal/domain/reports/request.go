package reports

// SummaryRequest represents a request for summary report
type SummaryRequest struct {
	CompanyID     int64  `form:"company_id" validate:"required,gt=0"`
	Year          int    `form:"year" validate:"required,gt=2000"`
	Months        string `form:"months"`                                  // Optional: "1,2,3" or empty for all months
	BudgetVersion int    `form:"budget_version" validate:"required,gte=1"` // Required: budget data version to compare against
}
