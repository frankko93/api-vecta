package reports

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gmhafiz/go8/internal/domain/data"
)

type UseCase interface {
	GetSummary(ctx context.Context, req *SummaryRequest) (*SummaryReport, error)
	SaveReport(ctx context.Context, req *SaveReportRequest, userID int64) (*SavedReport, error)
	ListSavedReports(ctx context.Context, companyID int64, year int) ([]*SavedReport, error)
	CompareReports(ctx context.Context, reportIDs []int64) (*CompareReportsResponse, error)
}

type useCase struct {
	repo       Repository
	calculator *Calculator
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{
		repo:       repo,
		calculator: NewCalculator(),
	}
}

func (uc *useCase) GetSummary(ctx context.Context, req *SummaryRequest) (*SummaryReport, error) {
	// Get company name
	companyName, err := uc.repo.GetCompanyName(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}

	// Get all data for the year
	pbrActual, err := uc.repo.GetPBRData(ctx, req.CompanyID, req.Year, "actual")
	if err != nil {
		return nil, err
	}

	pbrBudget, err := uc.repo.GetPBRData(ctx, req.CompanyID, req.Year, "budget")
	if err != nil {
		return nil, err
	}

	doreActual, err := uc.repo.GetDoreData(ctx, req.CompanyID, req.Year, "actual")
	if err != nil {
		return nil, err
	}

	doreBudget, err := uc.repo.GetDoreData(ctx, req.CompanyID, req.Year, "budget")
	if err != nil {
		return nil, err
	}

	opexActual, err := uc.repo.GetOPEXData(ctx, req.CompanyID, req.Year, "actual")
	if err != nil {
		return nil, err
	}

	opexBudget, err := uc.repo.GetOPEXData(ctx, req.CompanyID, req.Year, "budget")
	if err != nil {
		return nil, err
	}

	capexActual, err := uc.repo.GetCAPEXData(ctx, req.CompanyID, req.Year, "actual")
	if err != nil {
		return nil, err
	}

	capexBudget, err := uc.repo.GetCAPEXData(ctx, req.CompanyID, req.Year, "budget")
	if err != nil {
		return nil, err
	}

	financialActual, err := uc.repo.GetFinancialData(ctx, req.CompanyID, req.Year, "actual")
	if err != nil {
		return nil, err
	}

	financialBudget, err := uc.repo.GetFinancialData(ctx, req.CompanyID, req.Year, "budget")
	if err != nil {
		return nil, err
	}

	// Parse requested months (optional - if empty, returns all 12 months)
	monthsFilter := uc.parseMonthsFilter(req.Months)

	// Group data by month and calculate metrics
	months := uc.buildMonthlyData(
		req.Year,
		pbrActual, pbrBudget,
		doreActual, doreBudget,
		financialActual, financialBudget,
		opexActual, opexBudget,
		capexActual, capexBudget,
		monthsFilter,
	)

	return &SummaryReport{
		CompanyID:   req.CompanyID,
		CompanyName: companyName,
		Year:        req.Year,
		Months:      months,
	}, nil
}

func (uc *useCase) parseMonthsFilter(monthsStr string) map[int]bool {
	if monthsStr == "" {
		return nil // No filter, return all months
	}

	monthsFilter := make(map[int]bool)
	parts := strings.Split(monthsStr, ",")
	for _, p := range parts {
		month, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil && month >= 1 && month <= 12 {
			monthsFilter[month] = true
		}
	}

	return monthsFilter
}

func (uc *useCase) buildMonthlyData(
	year int,
	pbrActual, pbrBudget []*data.PBRData,
	doreActual, doreBudget []*data.DoreData,
	financialActual, financialBudget []*data.FinancialData,
	opexActual, opexBudget []*data.OPEXData,
	capexActual, capexBudget []*data.CAPEXData,
	monthsFilter map[int]bool,
) []MonthlyData {

	// Group data by month
	pbrActualByMonth := groupPBRByMonth(pbrActual)
	pbrBudgetByMonth := groupPBRByMonth(pbrBudget)
	doreActualByMonth := groupDoreByMonth(doreActual)
	doreBudgetByMonth := groupDoreByMonth(doreBudget)
	financialActualByMonth := groupFinancialByMonth(financialActual)
	financialBudgetByMonth := groupFinancialByMonth(financialBudget)
	opexActualByMonth := groupOPEXByMonth(opexActual)
	opexBudgetByMonth := groupOPEXByMonth(opexBudget)
	capexActualByMonth := groupCAPEXByMonth(capexActual)
	capexBudgetByMonth := groupCAPEXByMonth(capexBudget)

	var months []MonthlyData

	for month := 1; month <= 12; month++ {
		// Apply filter if specified (if nil, include all months)
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actualDataSet := uc.calculator.CalculateDataSet(
			pbrActualByMonth[month],
			doreActualByMonth[month],
			financialActualByMonth[month],
			opexActualByMonth[month],
			capexActualByMonth[month],
		)

		budgetDataSet := uc.calculator.CalculateDataSet(
			pbrBudgetByMonth[month],
			doreBudgetByMonth[month],
			financialBudgetByMonth[month],
			opexBudgetByMonth[month],
			capexBudgetByMonth[month],
		)

		// Calculate variance if both actual and budget exist
		var variance *VarianceData
		if actualDataSet != nil && budgetDataSet != nil {
			variance = uc.calculator.CalculateVarianceData(actualDataSet, budgetDataSet)
		}

		months = append(months, MonthlyData{
			Month:    monthKey,
			Actual:   actualDataSet,
			Budget:   budgetDataSet,
			Variance: variance,
		})
	}

	return months
}

// Helper functions to group data by month

func groupPBRByMonth(records []*data.PBRData) map[int]*data.PBRData {
	grouped := make(map[int]*data.PBRData)
	for _, r := range records {
		month := int(r.Date.Month())
		grouped[month] = r
	}
	return grouped
}

func groupDoreByMonth(records []*data.DoreData) map[int]*data.DoreData {
	grouped := make(map[int]*data.DoreData)
	for _, r := range records {
		month := int(r.Date.Month())
		grouped[month] = r
	}
	return grouped
}

func groupOPEXByMonth(records []*data.OPEXData) map[int][]*data.OPEXData {
	grouped := make(map[int][]*data.OPEXData)
	for _, r := range records {
		month := int(r.Date.Month())
		grouped[month] = append(grouped[month], r)
	}
	return grouped
}

func groupCAPEXByMonth(records []*data.CAPEXData) map[int][]*data.CAPEXData {
	grouped := make(map[int][]*data.CAPEXData)
	for _, r := range records {
		month := int(r.Date.Month())
		grouped[month] = append(grouped[month], r)
	}
	return grouped
}

func groupFinancialByMonth(records []*data.FinancialData) map[int]*data.FinancialData {
	grouped := make(map[int]*data.FinancialData)
	for _, r := range records {
		month := int(r.Date.Month())
		grouped[month] = r
	}
	return grouped
}
