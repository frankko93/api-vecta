package reports

import (
	"context"
	"log/slog"
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
	GetReportCompanyID(ctx context.Context, reportID int64) (int64, error)
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

	// Get company configuration (mining type, minerals)
	companyConfig, err := uc.repo.GetCompanyConfig(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}

	// Cross-file validation: log warnings but don't block the summary.
	// The report should work with whatever data is available.
	if err := validateCrossFile(ctx, uc.repo, req.CompanyID, req.Year); err != nil {
		slog.Warn("cross-file validation warning", "company_id", req.CompanyID, "year", req.Year, "warning", err.Error())
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

	// Compute data coverage (actual/budget months loaded)
	coverage := uc.buildCoverage(
		pbrActual, pbrBudget,
		doreActual, doreBudget,
		financialActual, financialBudget,
		opexActual, opexBudget,
		capexActual, capexBudget,
	)

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
		Config:      companyConfig,
		Months:      months,
		Coverage:    coverage,
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
	var ytdActual *DataSet
	var ytdBudget *DataSet

	for month := 1; month <= 12; month++ {
		// Apply filter if specified (if nil, include all months)
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actualHasData := pbrActualByMonth[month] != nil ||
			doreActualByMonth[month] != nil ||
			financialActualByMonth[month] != nil ||
			len(opexActualByMonth[month]) > 0 ||
			len(capexActualByMonth[month]) > 0

		budgetHasData := pbrBudgetByMonth[month] != nil ||
			doreBudgetByMonth[month] != nil ||
			financialBudgetByMonth[month] != nil ||
			len(opexBudgetByMonth[month]) > 0 ||
			len(capexBudgetByMonth[month]) > 0

		var actualDataSet *DataSet
		if actualHasData {
			actualDataSet = uc.calculator.CalculateDataSet(
				pbrActualByMonth[month],
				doreActualByMonth[month],
				financialActualByMonth[month],
				opexActualByMonth[month],
				capexActualByMonth[month],
			)
		}

		var budgetDataSet *DataSet
		if budgetHasData {
			budgetDataSet = uc.calculator.CalculateDataSet(
				pbrBudgetByMonth[month],
				doreBudgetByMonth[month],
				financialBudgetByMonth[month],
				opexBudgetByMonth[month],
				capexBudgetByMonth[month],
			)
		}

		// Calculate variance if both actual and budget exist
		var variance *VarianceData
		if actualDataSet != nil && budgetDataSet != nil {
			variance = uc.calculator.CalculateVarianceData(actualDataSet, budgetDataSet)
		}

		// Calculate YTD only up to the last loaded actual month
		var ytd *YTDData
		if actualHasData {
			// Get Dore data for current month (needed for gold credit calculation)
			actualDore := doreActualByMonth[month]
			budgetDore := doreBudgetByMonth[month]

			ytdActual = uc.calculator.AccumulateYTD(ytdActual, actualDataSet, actualDore)
			if budgetHasData {
				ytdBudget = uc.calculator.AccumulateYTD(ytdBudget, budgetDataSet, budgetDore)
			}

			var ytdVariance *VarianceData
			if ytdActual != nil && ytdBudget != nil {
				ytdVariance = uc.calculator.CalculateVarianceData(ytdActual, ytdBudget)
			}

			ytd = &YTDData{
				Actual:   ytdActual,
				Budget:   ytdBudget,
				Variance: ytdVariance,
			}
		}

		months = append(months, MonthlyData{
			Month:    monthKey,
			Actual:   actualDataSet,
			Budget:   budgetDataSet,
			Variance: variance,
			YTD:      ytd,
		})
	}

	return months
}

func (uc *useCase) buildCoverage(
	pbrActual, pbrBudget []*data.PBRData,
	doreActual, doreBudget []*data.DoreData,
	financialActual, financialBudget []*data.FinancialData,
	opexActual, opexBudget []*data.OPEXData,
	capexActual, capexBudget []*data.CAPEXData,
) *DataCoverage {
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

	actualMonths := collectMonthsWithData(
		pbrActualByMonth, doreActualByMonth, financialActualByMonth,
		opexActualByMonth, capexActualByMonth,
	)
	budgetMonths := collectMonthsWithData(
		pbrBudgetByMonth, doreBudgetByMonth, financialBudgetByMonth,
		opexBudgetByMonth, capexBudgetByMonth,
	)

	coverage := &DataCoverage{
		ActualMonths:      actualMonths,
		BudgetMonths:      budgetMonths,
		ActualLastMonth:   lastMonth(actualMonths),
		BudgetLastMonth:   lastMonth(budgetMonths),
		ActualIsPartial:   len(actualMonths) > 0 && len(actualMonths) < 12,
		BudgetIsPartial:   len(budgetMonths) > 0 && len(budgetMonths) < 12,
		HasAnyActual:      len(actualMonths) > 0,
		HasAnyBudget:      len(budgetMonths) > 0,
		HasCompleteActual: len(actualMonths) == 12,
		HasCompleteBudget: len(budgetMonths) == 12,
	}

	return coverage
}

func collectMonthsWithData(
	pbr map[int]*data.PBRData,
	dore map[int]*data.DoreData,
	financial map[int]*data.FinancialData,
	opex map[int][]*data.OPEXData,
	capex map[int][]*data.CAPEXData,
) []int {
	months := make([]int, 0, 12)
	for month := 1; month <= 12; month++ {
		hasData := pbr[month] != nil ||
			dore[month] != nil ||
			financial[month] != nil ||
			len(opex[month]) > 0 ||
			len(capex[month]) > 0
		if hasData {
			months = append(months, month)
		}
	}
	return months
}

func lastMonth(months []int) int {
	if len(months) == 0 {
		return 0
	}
	return months[len(months)-1]
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

// GetReportCompanyID returns the company ID for a saved report
func (uc *useCase) GetReportCompanyID(ctx context.Context, reportID int64) (int64, error) {
	return uc.repo.GetReportCompanyID(ctx, reportID)
}
