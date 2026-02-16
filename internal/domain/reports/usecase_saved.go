package reports

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// SaveReport saves a generated report as a scenario
func (uc *useCase) SaveReport(ctx context.Context, req *SaveReportRequest, userID int64) (*SavedReport, error) {
	// Budget version is required
	if req.BudgetVersion < 1 {
		return nil, fmt.Errorf("budget_version is required and must be >= 1")
	}

	// Generate complete report (all 12 months) using the specified budget version
	summaryReq := &SummaryRequest{
		CompanyID:     req.CompanyID,
		Year:          req.Year,
		BudgetVersion: req.BudgetVersion,
	}

	summary, err := uc.GetSummary(ctx, summaryReq)
	if err != nil {
		return nil, err
	}

	// Create saved report (with full year data)
	savedReport := &SavedReport{
		CompanyID:     req.CompanyID,
		Name:          req.Name,
		Description:   req.Description,
		Year:          req.Year,
		BudgetVersion: req.BudgetVersion,
		ReportData:    *summary,
		CreatedBy:     userID,
	}

	err = uc.repo.SaveReport(ctx, savedReport)
	if err != nil {
		return nil, err
	}

	return savedReport, nil
}

// ListSavedReports returns all saved reports for a company and year
func (uc *useCase) ListSavedReports(ctx context.Context, companyID int64, year int) ([]*SavedReport, error) {
	return uc.repo.ListSavedReports(ctx, companyID, year)
}

// CompareReports compares multiple saved reports side by side
func (uc *useCase) CompareReports(ctx context.Context, reportIDs []int64) (*CompareReportsResponse, error) {
	reports, err := uc.repo.GetSavedReportsByIDs(ctx, reportIDs)
	if err != nil {
		return nil, err
	}

	if len(reports) < 2 {
		return nil, fmt.Errorf("need at least 2 reports to compare")
	}

	// Build comparison (side by side data)
	comparison := buildComparison(reports)

	// Convert to slice of values for response
	reportSlice := make([]SavedReport, len(reports))
	for i, r := range reports {
		reportSlice[i] = *r
	}

	return &CompareReportsResponse{
		Reports:    reportSlice,
		Comparison: comparison,
	}, nil
}

// buildComparison creates side-by-side comparison data
func buildComparison(reports []*SavedReport) map[string]interface{} {
	// Extract key metrics from each report for easy comparison
	comparison := make(map[string]interface{})

	metrics := make([]map[string]interface{}, 0)

	for _, report := range reports {
		if len(report.ReportData.Months) > 0 {
			// Use first month as representative
			month := report.ReportData.Months[0]

			metrics = append(metrics, map[string]interface{}{
				"report_id":         report.ID,
				"report_name":       report.Name,
				"budget_version":    report.BudgetVersion,
				"ore_mined":         month.Actual.Mining.OreMinedT,
				"silver_production": month.Actual.Production.TotalProductionSilverOz,
				"gold_production":   month.Actual.Production.TotalProductionGoldOz,
				"production_costs":  month.Actual.Costs.ProductionBasedCosts,
				"nsr":               month.Actual.NSR.NetSmelterReturn,
				"production_margin": month.Actual.CAPEX.ProductionBasedMargin,
				"net_cash_flow":     month.Actual.CAPEX.PBRNetCashFlow,
				"cash_cost_per_oz":  month.Actual.CashCost.CashCostPerOzSilver,
				"aisc_per_oz":       month.Actual.CashCost.AISCPerOzSilver,
			})
		}
	}

	comparison["key_metrics"] = metrics
	comparison["count"] = len(reports)

	return comparison
}

// parseMonthsString converts "1,2,3" to []int{1,2,3}
func parseMonthsString(monthsStr string) []int {
	months := make([]int, 0)
	if monthsStr == "" {
		// Return all 12 months
		for i := 1; i <= 12; i++ {
			months = append(months, i)
		}
		return months
	}

	// Parse "1,2,3"
	parts := strings.Split(monthsStr, ",")
	for _, p := range parts {
		if m, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && m >= 1 && m <= 12 {
			months = append(months, m)
		}
	}

	if len(months) == 0 {
		for i := 1; i <= 12; i++ {
			months = append(months, i)
		}
	}

	return months
}
