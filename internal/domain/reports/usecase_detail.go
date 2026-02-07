package reports

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gmhafiz/go8/internal/domain/data"
)

// DetailUseCase interface for detailed reports
type DetailUseCase interface {
	GetPBRDetail(ctx context.Context, req *DetailRequest) (*PBRDetailReport, error)
	GetDoreDetail(ctx context.Context, req *DetailRequest) (*DoreDetailReport, error)
	GetOPEXDetail(ctx context.Context, req *DetailRequest) (*OPEXDetailReport, error)
	GetCAPEXDetail(ctx context.Context, req *DetailRequest) (*CAPEXDetailReport, error)
	// NOTE: GetFinancialDetail, GetProductionDetail, GetRevenueDetail removed
	// - Financial data is now in Summary/NSR and Summary/Costs
	// - Production data is now in PBR and Summary/Production
	// - Revenue data is now in Dore and Summary/NSR
}

// DetailRequest represents a request for detailed report
type DetailRequest struct {
	CompanyID int64  `form:"company_id" validate:"required,gt=0"`
	Year      int    `form:"year" validate:"required,gt=2000"`
	Months    string `form:"months"` // Optional: "1,2,3" or empty for all months
}

type detailUseCase struct {
	repo       Repository
	calculator *Calculator
}

func NewDetailUseCase(repo Repository) DetailUseCase {
	return &detailUseCase{
		repo:       repo,
		calculator: NewCalculator(),
	}
}

// GetPBRDetail returns detailed PBR report
func (uc *detailUseCase) GetPBRDetail(ctx context.Context, req *DetailRequest) (*PBRDetailReport, error) {
	companyName, err := uc.repo.GetCompanyName(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}

	companyConfig, err := uc.repo.GetCompanyConfig(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}

	pbrActual, err := uc.repo.GetPBRData(ctx, req.CompanyID, req.Year, "actual")
	if err != nil {
		return nil, err
	}

	pbrBudget, err := uc.repo.GetPBRData(ctx, req.CompanyID, req.Year, "budget")
	if err != nil {
		return nil, err
	}

	monthsFilter := uc.parseMonthsFilter(req.Months)
	months := uc.buildPBRMonthlyData(req.Year, pbrActual, pbrBudget, monthsFilter)

	return &PBRDetailReport{
		CompanyID:   req.CompanyID,
		CompanyName: companyName,
		Year:        req.Year,
		Config:      companyConfig,
		Months:      months,
	}, nil
}

// GetDoreDetail returns detailed Dore report
func (uc *detailUseCase) GetDoreDetail(ctx context.Context, req *DetailRequest) (*DoreDetailReport, error) {
	companyName, err := uc.repo.GetCompanyName(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}

	companyConfig, err := uc.repo.GetCompanyConfig(ctx, req.CompanyID)
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

	pbrActual, err := uc.repo.GetPBRData(ctx, req.CompanyID, req.Year, "actual")
	if err != nil {
		return nil, err
	}

	pbrBudget, err := uc.repo.GetPBRData(ctx, req.CompanyID, req.Year, "budget")
	if err != nil {
		return nil, err
	}

	monthsFilter := uc.parseMonthsFilter(req.Months)
	months := uc.buildDoreMonthlyData(req.Year, doreActual, doreBudget, pbrActual, pbrBudget, monthsFilter)

	return &DoreDetailReport{
		CompanyID:   req.CompanyID,
		CompanyName: companyName,
		Year:        req.Year,
		Config:      companyConfig,
		Months:      months,
	}, nil
}

// GetOPEXDetail returns detailed OPEX report
func (uc *detailUseCase) GetOPEXDetail(ctx context.Context, req *DetailRequest) (*OPEXDetailReport, error) {
	companyName, err := uc.repo.GetCompanyName(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}

	companyConfig, err := uc.repo.GetCompanyConfig(ctx, req.CompanyID)
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

	monthsFilter := uc.parseMonthsFilter(req.Months)
	months, byCostCenter, bySubcategory, byExpenseType := uc.buildOPEXMonthlyData(req.Year, opexActual, opexBudget, monthsFilter)

	return &OPEXDetailReport{
		CompanyID:     req.CompanyID,
		CompanyName:   companyName,
		Year:          req.Year,
		Config:        companyConfig,
		Months:        months,
		ByCostCenter:  byCostCenter,
		BySubcategory: bySubcategory,
		ByExpenseType: byExpenseType,
	}, nil
}

// GetCAPEXDetail returns detailed CAPEX report
func (uc *detailUseCase) GetCAPEXDetail(ctx context.Context, req *DetailRequest) (*CAPEXDetailReport, error) {
	companyName, err := uc.repo.GetCompanyName(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}

	companyConfig, err := uc.repo.GetCompanyConfig(ctx, req.CompanyID)
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

	monthsFilter := uc.parseMonthsFilter(req.Months)
	months, byType, byCategory := uc.buildCAPEXMonthlyData(req.Year, capexActual, capexBudget, monthsFilter)

	return &CAPEXDetailReport{
		CompanyID:   req.CompanyID,
		CompanyName: companyName,
		Year:        req.Year,
		Config:      companyConfig,
		Months:      months,
		ByType:      byType,
		ByCategory:  byCategory,
	}, nil
}

// NOTE: GetFinancialDetail, GetProductionDetail, GetRevenueDetail removed
// - Financial data is now in Summary/NSR and Summary/Costs
// - Production data is now in PBR and Summary/Production
// - Revenue data is now in Dore and Summary/NSR

// Helper methods

func (uc *detailUseCase) parseMonthsFilter(monthsStr string) map[int]bool {
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

// buildPBRMonthlyData builds PBR monthly data with variances
func (uc *detailUseCase) buildPBRMonthlyData(
	year int,
	pbrActual, pbrBudget []*data.PBRData,
	monthsFilter map[int]bool,
) []PBRMonthlyData {
	pbrActualByMonth := groupPBRByMonth(pbrActual)
	pbrBudgetByMonth := groupPBRByMonth(pbrBudget)

	var months []PBRMonthlyData

	for month := 1; month <= 12; month++ {
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actual := uc.buildPBRDetail(pbrActualByMonth[month])
		budget := uc.buildPBRDetail(pbrBudgetByMonth[month])

		var variance *PBRVariance
		if actual != nil && budget != nil {
			variance = uc.calculatePBRVariance(actual, budget)
		}

		months = append(months, PBRMonthlyData{
			Month:    monthKey,
			Actual:   actual,
			Budget:   budget,
			Variance: variance,
		})
	}

	return months
}

func (uc *detailUseCase) buildPBRDetail(pbr *data.PBRData) *PBRDetail {
	if pbr == nil {
		return nil
	}

	// Calculate ratios
	var wasteOreRatio float64
	if pbr.OreMinedT > 0 {
		wasteOreRatio = pbr.WasteMinedT / pbr.OreMinedT
	}
	totalMoved := pbr.OreMinedT + pbr.WasteMinedT

	// Calculate production
	silverOz := pbr.FeedGradeSilverGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateSilverPct / 100) / 31.1035
	goldOz := pbr.FeedGradeGoldGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateGoldPct / 100) / 31.1035

	return &PBRDetail{
		// Mining - Ore breakdown
		OpenPitOreT:     pbr.OpenPitOreT,
		UndergroundOreT: pbr.UndergroundOreT,
		OreMinedT:       pbr.OreMinedT,

		// Mining - Waste and ratios
		WasteMinedT:    pbr.WasteMinedT,
		StrippingRatio: pbr.StrippingRatio,
		WasteOreRatio:  wasteOreRatio,
		TotalMoved:     totalMoved,

		// Mining - Grades
		MiningGradeSilverGpt:      pbr.MiningGradeSilverGpt,
		MiningGradeGoldGpt:        pbr.MiningGradeGoldGpt,
		OpenPitGradeSilverGpt:     pbr.OpenPitGradeSilverGpt,
		UndergroundGradeSilverGpt: pbr.UndergroundGradeSilverGpt,
		OpenPitGradeGoldGpt:       pbr.OpenPitGradeGoldGpt,
		UndergroundGradeGoldGpt:   pbr.UndergroundGradeGoldGpt,

		// Developments breakdown
		PrimaryDevelopmentM:       pbr.PrimaryDevelopmentM,
		SecondaryDevelopmentOpexM: pbr.SecondaryDevelopmentOpexM,
		ExpansionaryDevelopmentM:  pbr.ExpansionaryDevelopmentM,
		DevelopmentsM:             pbr.DevelopmentsM,

		// Processing
		TotalTonnesProcessed:  pbr.TotalTonnesProcessed,
		FeedGradeSilverGpt:    pbr.FeedGradeSilverGpt,
		FeedGradeGoldGpt:      pbr.FeedGradeGoldGpt,
		RecoveryRateSilverPct: pbr.RecoveryRateSilverPct,
		RecoveryRateGoldPct:   pbr.RecoveryRateGoldPct,

		// Production (calculated)
		TotalProductionSilverOz: silverOz,
		TotalProductionGoldOz:   goldOz,

		// Headcount
		FullTimeEmployees: pbr.FullTimeEmployees,
		Contractors:       pbr.Contractors,
		TotalHeadcount:    pbr.TotalHeadcount,

		HasData: true,
	}
}

func (uc *detailUseCase) calculatePBRVariance(actual, budget *PBRDetail) *PBRVariance {
	return &PBRVariance{
		// Mining - Ore breakdown
		OpenPitOreT:     VarianceMetric{Actual: actual.OpenPitOreT, Budget: budget.OpenPitOreT, Variance: actual.OpenPitOreT - budget.OpenPitOreT, VariancePct: calculateVariancePct(actual.OpenPitOreT, budget.OpenPitOreT)},
		UndergroundOreT: VarianceMetric{Actual: actual.UndergroundOreT, Budget: budget.UndergroundOreT, Variance: actual.UndergroundOreT - budget.UndergroundOreT, VariancePct: calculateVariancePct(actual.UndergroundOreT, budget.UndergroundOreT)},
		OreMinedT:       VarianceMetric{Actual: actual.OreMinedT, Budget: budget.OreMinedT, Variance: actual.OreMinedT - budget.OreMinedT, VariancePct: calculateVariancePct(actual.OreMinedT, budget.OreMinedT)},

		// Mining - Waste and ratios
		WasteMinedT:    VarianceMetric{Actual: actual.WasteMinedT, Budget: budget.WasteMinedT, Variance: actual.WasteMinedT - budget.WasteMinedT, VariancePct: calculateVariancePct(actual.WasteMinedT, budget.WasteMinedT)},
		StrippingRatio: VarianceMetric{Actual: actual.StrippingRatio, Budget: budget.StrippingRatio, Variance: actual.StrippingRatio - budget.StrippingRatio, VariancePct: calculateVariancePct(actual.StrippingRatio, budget.StrippingRatio)},
		WasteOreRatio:  VarianceMetric{Actual: actual.WasteOreRatio, Budget: budget.WasteOreRatio, Variance: actual.WasteOreRatio - budget.WasteOreRatio, VariancePct: calculateVariancePct(actual.WasteOreRatio, budget.WasteOreRatio)},
		TotalMoved:     VarianceMetric{Actual: actual.TotalMoved, Budget: budget.TotalMoved, Variance: actual.TotalMoved - budget.TotalMoved, VariancePct: calculateVariancePct(actual.TotalMoved, budget.TotalMoved)},

		// Mining - Grades
		MiningGradeSilverGpt:      VarianceMetric{Actual: actual.MiningGradeSilverGpt, Budget: budget.MiningGradeSilverGpt, Variance: actual.MiningGradeSilverGpt - budget.MiningGradeSilverGpt, VariancePct: calculateVariancePct(actual.MiningGradeSilverGpt, budget.MiningGradeSilverGpt)},
		MiningGradeGoldGpt:        VarianceMetric{Actual: actual.MiningGradeGoldGpt, Budget: budget.MiningGradeGoldGpt, Variance: actual.MiningGradeGoldGpt - budget.MiningGradeGoldGpt, VariancePct: calculateVariancePct(actual.MiningGradeGoldGpt, budget.MiningGradeGoldGpt)},
		OpenPitGradeSilverGpt:     VarianceMetric{Actual: actual.OpenPitGradeSilverGpt, Budget: budget.OpenPitGradeSilverGpt, Variance: actual.OpenPitGradeSilverGpt - budget.OpenPitGradeSilverGpt, VariancePct: calculateVariancePct(actual.OpenPitGradeSilverGpt, budget.OpenPitGradeSilverGpt)},
		UndergroundGradeSilverGpt: VarianceMetric{Actual: actual.UndergroundGradeSilverGpt, Budget: budget.UndergroundGradeSilverGpt, Variance: actual.UndergroundGradeSilverGpt - budget.UndergroundGradeSilverGpt, VariancePct: calculateVariancePct(actual.UndergroundGradeSilverGpt, budget.UndergroundGradeSilverGpt)},
		OpenPitGradeGoldGpt:       VarianceMetric{Actual: actual.OpenPitGradeGoldGpt, Budget: budget.OpenPitGradeGoldGpt, Variance: actual.OpenPitGradeGoldGpt - budget.OpenPitGradeGoldGpt, VariancePct: calculateVariancePct(actual.OpenPitGradeGoldGpt, budget.OpenPitGradeGoldGpt)},
		UndergroundGradeGoldGpt:   VarianceMetric{Actual: actual.UndergroundGradeGoldGpt, Budget: budget.UndergroundGradeGoldGpt, Variance: actual.UndergroundGradeGoldGpt - budget.UndergroundGradeGoldGpt, VariancePct: calculateVariancePct(actual.UndergroundGradeGoldGpt, budget.UndergroundGradeGoldGpt)},

		// Developments breakdown
		PrimaryDevelopmentM:       VarianceMetric{Actual: actual.PrimaryDevelopmentM, Budget: budget.PrimaryDevelopmentM, Variance: actual.PrimaryDevelopmentM - budget.PrimaryDevelopmentM, VariancePct: calculateVariancePct(actual.PrimaryDevelopmentM, budget.PrimaryDevelopmentM)},
		SecondaryDevelopmentOpexM: VarianceMetric{Actual: actual.SecondaryDevelopmentOpexM, Budget: budget.SecondaryDevelopmentOpexM, Variance: actual.SecondaryDevelopmentOpexM - budget.SecondaryDevelopmentOpexM, VariancePct: calculateVariancePct(actual.SecondaryDevelopmentOpexM, budget.SecondaryDevelopmentOpexM)},
		ExpansionaryDevelopmentM:  VarianceMetric{Actual: actual.ExpansionaryDevelopmentM, Budget: budget.ExpansionaryDevelopmentM, Variance: actual.ExpansionaryDevelopmentM - budget.ExpansionaryDevelopmentM, VariancePct: calculateVariancePct(actual.ExpansionaryDevelopmentM, budget.ExpansionaryDevelopmentM)},
		DevelopmentsM:             VarianceMetric{Actual: actual.DevelopmentsM, Budget: budget.DevelopmentsM, Variance: actual.DevelopmentsM - budget.DevelopmentsM, VariancePct: calculateVariancePct(actual.DevelopmentsM, budget.DevelopmentsM)},

		// Processing
		TotalTonnesProcessed:  VarianceMetric{Actual: actual.TotalTonnesProcessed, Budget: budget.TotalTonnesProcessed, Variance: actual.TotalTonnesProcessed - budget.TotalTonnesProcessed, VariancePct: calculateVariancePct(actual.TotalTonnesProcessed, budget.TotalTonnesProcessed)},
		FeedGradeSilverGpt:    VarianceMetric{Actual: actual.FeedGradeSilverGpt, Budget: budget.FeedGradeSilverGpt, Variance: actual.FeedGradeSilverGpt - budget.FeedGradeSilverGpt, VariancePct: calculateVariancePct(actual.FeedGradeSilverGpt, budget.FeedGradeSilverGpt)},
		FeedGradeGoldGpt:      VarianceMetric{Actual: actual.FeedGradeGoldGpt, Budget: budget.FeedGradeGoldGpt, Variance: actual.FeedGradeGoldGpt - budget.FeedGradeGoldGpt, VariancePct: calculateVariancePct(actual.FeedGradeGoldGpt, budget.FeedGradeGoldGpt)},
		RecoveryRateSilverPct: VarianceMetric{Actual: actual.RecoveryRateSilverPct, Budget: budget.RecoveryRateSilverPct, Variance: actual.RecoveryRateSilverPct - budget.RecoveryRateSilverPct, VariancePct: calculateVariancePct(actual.RecoveryRateSilverPct, budget.RecoveryRateSilverPct)},
		RecoveryRateGoldPct:   VarianceMetric{Actual: actual.RecoveryRateGoldPct, Budget: budget.RecoveryRateGoldPct, Variance: actual.RecoveryRateGoldPct - budget.RecoveryRateGoldPct, VariancePct: calculateVariancePct(actual.RecoveryRateGoldPct, budget.RecoveryRateGoldPct)},

		// Production
		TotalProductionSilverOz: VarianceMetric{Actual: actual.TotalProductionSilverOz, Budget: budget.TotalProductionSilverOz, Variance: actual.TotalProductionSilverOz - budget.TotalProductionSilverOz, VariancePct: calculateVariancePct(actual.TotalProductionSilverOz, budget.TotalProductionSilverOz)},
		TotalProductionGoldOz:   VarianceMetric{Actual: actual.TotalProductionGoldOz, Budget: budget.TotalProductionGoldOz, Variance: actual.TotalProductionGoldOz - budget.TotalProductionGoldOz, VariancePct: calculateVariancePct(actual.TotalProductionGoldOz, budget.TotalProductionGoldOz)},

		// Headcount
		FullTimeEmployees: VarianceMetric{Actual: float64(actual.FullTimeEmployees), Budget: float64(budget.FullTimeEmployees), Variance: float64(actual.FullTimeEmployees - budget.FullTimeEmployees), VariancePct: calculateVariancePct(float64(actual.FullTimeEmployees), float64(budget.FullTimeEmployees))},
		Contractors:       VarianceMetric{Actual: float64(actual.Contractors), Budget: float64(budget.Contractors), Variance: float64(actual.Contractors - budget.Contractors), VariancePct: calculateVariancePct(float64(actual.Contractors), float64(budget.Contractors))},
		TotalHeadcount:    VarianceMetric{Actual: float64(actual.TotalHeadcount), Budget: float64(budget.TotalHeadcount), Variance: float64(actual.TotalHeadcount - budget.TotalHeadcount), VariancePct: calculateVariancePct(float64(actual.TotalHeadcount), float64(budget.TotalHeadcount))},
	}
}

// buildDoreMonthlyData builds Dore monthly data with variances
func (uc *detailUseCase) buildDoreMonthlyData(
	year int,
	doreActual, doreBudget []*data.DoreData,
	pbrActual, pbrBudget []*data.PBRData,
	monthsFilter map[int]bool,
) []DoreMonthlyData {
	doreActualByMonth := groupDoreByMonth(doreActual)
	doreBudgetByMonth := groupDoreByMonth(doreBudget)
	pbrActualByMonth := groupPBRByMonth(pbrActual)
	pbrBudgetByMonth := groupPBRByMonth(pbrBudget)

	var months []DoreMonthlyData

	for month := 1; month <= 12; month++ {
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actual := uc.buildDoreDetail(doreActualByMonth[month], pbrActualByMonth[month])
		budget := uc.buildDoreDetail(doreBudgetByMonth[month], pbrBudgetByMonth[month])

		var variance *DoreVariance
		if actual != nil && budget != nil {
			variance = uc.calculateDoreVariance(actual, budget)
		}

		months = append(months, DoreMonthlyData{
			Month:    monthKey,
			Actual:   actual,
			Budget:   budget,
			Variance: variance,
		})
	}

	return months
}

func (uc *detailUseCase) buildDoreDetail(dore *data.DoreData, pbr *data.PBRData) *DoreDetail {
	if dore == nil {
		return nil
	}

	// Metal in dore (before adjustments)
	metalInDoreSilverOz := dore.DoreProducedOz * (dore.SilverGradePct / 100)
	metalInDoreGoldOz := dore.DoreProducedOz * (dore.GoldGradePct / 100)

	// Metal adjusted (after adjustments)
	metalAdjustedSilverOz := metalInDoreSilverOz + dore.SilverAdjustmentOz
	metalAdjustedGoldOz := metalInDoreGoldOz + dore.GoldAdjustmentOz

	// Deductions
	deductionsSilverOz := metalAdjustedSilverOz * (dore.AgDeductionsPct / 100)
	deductionsGoldOz := metalAdjustedGoldOz * (dore.AuDeductionsPct / 100)

	// Payable metal (after deductions)
	payableSilverOz := metalAdjustedSilverOz - deductionsSilverOz
	payableGoldOz := metalAdjustedGoldOz - deductionsGoldOz

	// Gross revenue
	grossRevenueSilver := payableSilverOz * dore.RealizedPriceSilver
	grossRevenueGold := payableGoldOz * dore.RealizedPriceGold
	grossRevenueTotal := grossRevenueSilver + grossRevenueGold

	// Charges
	totalCharges := dore.TreatmentCharge + dore.RefiningDeductionsAu

	// NSR Dore
	nsrDore := grossRevenueTotal - totalCharges

	return &DoreDetail{
		DoreProducedOz:        dore.DoreProducedOz,
		SilverGradePct:        dore.SilverGradePct,
		GoldGradePct:          dore.GoldGradePct,
		MetalInDoreSilverOz:   metalInDoreSilverOz,
		MetalInDoreGoldOz:     metalInDoreGoldOz,
		SilverAdjustmentOz:    dore.SilverAdjustmentOz,
		GoldAdjustmentOz:      dore.GoldAdjustmentOz,
		MetalAdjustedSilverOz: metalAdjustedSilverOz,
		MetalAdjustedGoldOz:   metalAdjustedGoldOz,
		AgDeductionsPct:       dore.AgDeductionsPct,
		AuDeductionsPct:       dore.AuDeductionsPct,
		DeductionsSilverOz:    deductionsSilverOz,
		DeductionsGoldOz:      deductionsGoldOz,
		PayableSilverOz:       payableSilverOz,
		PayableGoldOz:         payableGoldOz,
		PBRPriceSilver:        dore.PBRPriceSilver,
		PBRPriceGold:          dore.PBRPriceGold,
		RealizedPriceSilver:   dore.RealizedPriceSilver,
		RealizedPriceGold:     dore.RealizedPriceGold,
		GrossRevenueSilver:    grossRevenueSilver,
		GrossRevenueGold:      grossRevenueGold,
		GrossRevenueTotal:     grossRevenueTotal,
		TreatmentCharge:       dore.TreatmentCharge,
		RefiningDeductionsAu:  dore.RefiningDeductionsAu,
		TotalCharges:          totalCharges,
		NSRDore:               nsrDore,
		HasData:               true,
	}
}

func (uc *detailUseCase) calculateDoreVariance(actual, budget *DoreDetail) *DoreVariance {
	return &DoreVariance{
		DoreProducedOz:        VarianceMetric{Actual: actual.DoreProducedOz, Budget: budget.DoreProducedOz, Variance: actual.DoreProducedOz - budget.DoreProducedOz, VariancePct: calculateVariancePct(actual.DoreProducedOz, budget.DoreProducedOz)},
		SilverGradePct:        VarianceMetric{Actual: actual.SilverGradePct, Budget: budget.SilverGradePct, Variance: actual.SilverGradePct - budget.SilverGradePct, VariancePct: calculateVariancePct(actual.SilverGradePct, budget.SilverGradePct)},
		GoldGradePct:          VarianceMetric{Actual: actual.GoldGradePct, Budget: budget.GoldGradePct, Variance: actual.GoldGradePct - budget.GoldGradePct, VariancePct: calculateVariancePct(actual.GoldGradePct, budget.GoldGradePct)},
		MetalInDoreSilverOz:   VarianceMetric{Actual: actual.MetalInDoreSilverOz, Budget: budget.MetalInDoreSilverOz, Variance: actual.MetalInDoreSilverOz - budget.MetalInDoreSilverOz, VariancePct: calculateVariancePct(actual.MetalInDoreSilverOz, budget.MetalInDoreSilverOz)},
		MetalInDoreGoldOz:     VarianceMetric{Actual: actual.MetalInDoreGoldOz, Budget: budget.MetalInDoreGoldOz, Variance: actual.MetalInDoreGoldOz - budget.MetalInDoreGoldOz, VariancePct: calculateVariancePct(actual.MetalInDoreGoldOz, budget.MetalInDoreGoldOz)},
		SilverAdjustmentOz:    VarianceMetric{Actual: actual.SilverAdjustmentOz, Budget: budget.SilverAdjustmentOz, Variance: actual.SilverAdjustmentOz - budget.SilverAdjustmentOz, VariancePct: calculateVariancePct(actual.SilverAdjustmentOz, budget.SilverAdjustmentOz)},
		GoldAdjustmentOz:      VarianceMetric{Actual: actual.GoldAdjustmentOz, Budget: budget.GoldAdjustmentOz, Variance: actual.GoldAdjustmentOz - budget.GoldAdjustmentOz, VariancePct: calculateVariancePct(actual.GoldAdjustmentOz, budget.GoldAdjustmentOz)},
		MetalAdjustedSilverOz: VarianceMetric{Actual: actual.MetalAdjustedSilverOz, Budget: budget.MetalAdjustedSilverOz, Variance: actual.MetalAdjustedSilverOz - budget.MetalAdjustedSilverOz, VariancePct: calculateVariancePct(actual.MetalAdjustedSilverOz, budget.MetalAdjustedSilverOz)},
		MetalAdjustedGoldOz:   VarianceMetric{Actual: actual.MetalAdjustedGoldOz, Budget: budget.MetalAdjustedGoldOz, Variance: actual.MetalAdjustedGoldOz - budget.MetalAdjustedGoldOz, VariancePct: calculateVariancePct(actual.MetalAdjustedGoldOz, budget.MetalAdjustedGoldOz)},
		DeductionsSilverOz:    VarianceMetric{Actual: actual.DeductionsSilverOz, Budget: budget.DeductionsSilverOz, Variance: actual.DeductionsSilverOz - budget.DeductionsSilverOz, VariancePct: calculateVariancePct(actual.DeductionsSilverOz, budget.DeductionsSilverOz)},
		DeductionsGoldOz:      VarianceMetric{Actual: actual.DeductionsGoldOz, Budget: budget.DeductionsGoldOz, Variance: actual.DeductionsGoldOz - budget.DeductionsGoldOz, VariancePct: calculateVariancePct(actual.DeductionsGoldOz, budget.DeductionsGoldOz)},
		PayableSilverOz:       VarianceMetric{Actual: actual.PayableSilverOz, Budget: budget.PayableSilverOz, Variance: actual.PayableSilverOz - budget.PayableSilverOz, VariancePct: calculateVariancePct(actual.PayableSilverOz, budget.PayableSilverOz)},
		PayableGoldOz:         VarianceMetric{Actual: actual.PayableGoldOz, Budget: budget.PayableGoldOz, Variance: actual.PayableGoldOz - budget.PayableGoldOz, VariancePct: calculateVariancePct(actual.PayableGoldOz, budget.PayableGoldOz)},
		GrossRevenueSilver:    VarianceMetric{Actual: actual.GrossRevenueSilver, Budget: budget.GrossRevenueSilver, Variance: actual.GrossRevenueSilver - budget.GrossRevenueSilver, VariancePct: calculateVariancePct(actual.GrossRevenueSilver, budget.GrossRevenueSilver)},
		GrossRevenueGold:      VarianceMetric{Actual: actual.GrossRevenueGold, Budget: budget.GrossRevenueGold, Variance: actual.GrossRevenueGold - budget.GrossRevenueGold, VariancePct: calculateVariancePct(actual.GrossRevenueGold, budget.GrossRevenueGold)},
		GrossRevenueTotal:     VarianceMetric{Actual: actual.GrossRevenueTotal, Budget: budget.GrossRevenueTotal, Variance: actual.GrossRevenueTotal - budget.GrossRevenueTotal, VariancePct: calculateVariancePct(actual.GrossRevenueTotal, budget.GrossRevenueTotal)},
		TreatmentCharge:       VarianceMetric{Actual: actual.TreatmentCharge, Budget: budget.TreatmentCharge, Variance: actual.TreatmentCharge - budget.TreatmentCharge, VariancePct: calculateVariancePct(actual.TreatmentCharge, budget.TreatmentCharge)},
		RefiningDeductionsAu:  VarianceMetric{Actual: actual.RefiningDeductionsAu, Budget: budget.RefiningDeductionsAu, Variance: actual.RefiningDeductionsAu - budget.RefiningDeductionsAu, VariancePct: calculateVariancePct(actual.RefiningDeductionsAu, budget.RefiningDeductionsAu)},
		TotalCharges:          VarianceMetric{Actual: actual.TotalCharges, Budget: budget.TotalCharges, Variance: actual.TotalCharges - budget.TotalCharges, VariancePct: calculateVariancePct(actual.TotalCharges, budget.TotalCharges)},
		NSRDore:               VarianceMetric{Actual: actual.NSRDore, Budget: budget.NSRDore, Variance: actual.NSRDore - budget.NSRDore, VariancePct: calculateVariancePct(actual.NSRDore, budget.NSRDore)},
	}
}

// buildOPEXMonthlyData builds OPEX monthly data with variances and aggregations
func (uc *detailUseCase) buildOPEXMonthlyData(
	year int,
	opexActual, opexBudget []*data.OPEXData,
	monthsFilter map[int]bool,
) ([]OPEXMonthlyData, map[string]OPEXCostCenterData, map[string]OPEXSubcategoryData, map[string]OPEXExpenseTypeData) {
	opexActualByMonth := groupOPEXByMonth(opexActual)
	opexBudgetByMonth := groupOPEXByMonth(opexBudget)

	var months []OPEXMonthlyData
	costCenterTotals := make(map[string]struct{ Actual, Budget float64 })

	// Track subcategory totals with their cost center
	subcategoryTotals := make(map[string]struct {
		CostCenter string
		Actual     float64
		Budget     float64
	})

	// Track expense type totals
	expenseTypeTotals := make(map[string]struct{ Actual, Budget float64 })

	for month := 1; month <= 12; month++ {
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actual := uc.buildOPEXDetail(opexActualByMonth[month])
		budget := uc.buildOPEXDetail(opexBudgetByMonth[month])

		// Aggregate by cost center
		if actual != nil {
			costCenterTotals["Mine"] = struct{ Actual, Budget float64 }{costCenterTotals["Mine"].Actual + actual.Mine, costCenterTotals["Mine"].Budget}
			costCenterTotals["Processing"] = struct{ Actual, Budget float64 }{costCenterTotals["Processing"].Actual + actual.Processing, costCenterTotals["Processing"].Budget}
			costCenterTotals["G&A"] = struct{ Actual, Budget float64 }{costCenterTotals["G&A"].Actual + actual.GA, costCenterTotals["G&A"].Budget}
			costCenterTotals["Transport & Shipping"] = struct{ Actual, Budget float64 }{costCenterTotals["Transport & Shipping"].Actual + actual.TransportShipping, costCenterTotals["Transport & Shipping"].Budget}
		}
		if budget != nil {
			costCenterTotals["Mine"] = struct{ Actual, Budget float64 }{costCenterTotals["Mine"].Actual, costCenterTotals["Mine"].Budget + budget.Mine}
			costCenterTotals["Processing"] = struct{ Actual, Budget float64 }{costCenterTotals["Processing"].Actual, costCenterTotals["Processing"].Budget + budget.Processing}
			costCenterTotals["G&A"] = struct{ Actual, Budget float64 }{costCenterTotals["G&A"].Actual, costCenterTotals["G&A"].Budget + budget.GA}
			costCenterTotals["Transport & Shipping"] = struct{ Actual, Budget float64 }{costCenterTotals["Transport & Shipping"].Actual, costCenterTotals["Transport & Shipping"].Budget + budget.TransportShipping}
		}

		// Aggregate by subcategory from raw data (for actual)
		for _, opex := range opexActualByMonth[month] {
			key := opex.Subcategory
			entry := subcategoryTotals[key]
			entry.CostCenter = opex.CostCenter
			entry.Actual += opex.Amount
			subcategoryTotals[key] = entry

			// Aggregate by expense type
			etEntry := expenseTypeTotals[opex.ExpenseType]
			etEntry.Actual += opex.Amount
			expenseTypeTotals[opex.ExpenseType] = etEntry
		}

		// Aggregate by subcategory from raw data (for budget)
		for _, opex := range opexBudgetByMonth[month] {
			key := opex.Subcategory
			entry := subcategoryTotals[key]
			entry.CostCenter = opex.CostCenter
			entry.Budget += opex.Amount
			subcategoryTotals[key] = entry

			// Aggregate by expense type
			etEntry := expenseTypeTotals[opex.ExpenseType]
			etEntry.Budget += opex.Amount
			expenseTypeTotals[opex.ExpenseType] = etEntry
		}

		var variance *OPEXVariance
		if actual != nil && budget != nil {
			variance = uc.calculateOPEXVariance(actual, budget)
		}

		months = append(months, OPEXMonthlyData{
			Month:    monthKey,
			Actual:   actual,
			Budget:   budget,
			Variance: variance,
		})
	}

	// Build cost center aggregations
	byCostCenter := make(map[string]OPEXCostCenterData)
	for center, totals := range costCenterTotals {
		byCostCenter[center] = OPEXCostCenterData{
			CostCenter: center,
			Actual:     totals.Actual,
			Budget:     totals.Budget,
			Variance:   VarianceMetric{Actual: totals.Actual, Budget: totals.Budget, Variance: totals.Actual - totals.Budget, VariancePct: calculateVariancePct(totals.Actual, totals.Budget)},
		}
	}

	// Build subcategory aggregations
	bySubcategory := make(map[string]OPEXSubcategoryData)
	for subcategory, totals := range subcategoryTotals {
		bySubcategory[subcategory] = OPEXSubcategoryData{
			Subcategory: subcategory,
			CostCenter:  totals.CostCenter,
			Actual:      totals.Actual,
			Budget:      totals.Budget,
			Variance:    VarianceMetric{Actual: totals.Actual, Budget: totals.Budget, Variance: totals.Actual - totals.Budget, VariancePct: calculateVariancePct(totals.Actual, totals.Budget)},
		}
	}

	// Build expense type aggregations
	byExpenseType := make(map[string]OPEXExpenseTypeData)
	for expenseType, totals := range expenseTypeTotals {
		byExpenseType[expenseType] = OPEXExpenseTypeData{
			ExpenseType: expenseType,
			Actual:      totals.Actual,
			Budget:      totals.Budget,
			Variance:    VarianceMetric{Actual: totals.Actual, Budget: totals.Budget, Variance: totals.Actual - totals.Budget, VariancePct: calculateVariancePct(totals.Actual, totals.Budget)},
		}
	}

	return months, byCostCenter, bySubcategory, byExpenseType
}

func (uc *detailUseCase) buildOPEXDetail(opexList []*data.OPEXData) *OPEXDetail {
	if len(opexList) == 0 {
		return nil
	}

	var mine, processing, ga, transport, inventory float64
	bySubcategory := make(map[string]float64)
	byExpenseType := make(map[string]float64)

	for _, opex := range opexList {
		// Track by expense type
		byExpenseType[opex.ExpenseType] += opex.Amount

		// Inventory variations handling
		if opex.Subcategory == "Inventory Variation" || opex.Subcategory == "Stockpile/WIP" || opex.Subcategory == "Inventory Variations" {
			inventory += opex.Amount
			bySubcategory[opex.Subcategory] += opex.Amount
			continue
		}

		bySubcategory[opex.Subcategory] += opex.Amount

		switch opex.CostCenter {
		case "Mine":
			mine += opex.Amount
		case "Processing":
			processing += opex.Amount
		case "G&A":
			ga += opex.Amount
		case "Transport & Shipping":
			transport += opex.Amount
		}
	}

	total := mine + processing + ga + transport + inventory

	return &OPEXDetail{
		Mine:                mine,
		Processing:          processing,
		GA:                  ga,
		TransportShipping:   transport,
		InventoryVariations: inventory,
		Total:               total,
		BySubcategory:       bySubcategory,
		ByExpenseType:       byExpenseType,
		HasData:             true,
	}
}

func (uc *detailUseCase) calculateOPEXVariance(actual, budget *OPEXDetail) *OPEXVariance {
	return &OPEXVariance{
		Mine:                VarianceMetric{Actual: actual.Mine, Budget: budget.Mine, Variance: actual.Mine - budget.Mine, VariancePct: calculateVariancePct(actual.Mine, budget.Mine)},
		Processing:          VarianceMetric{Actual: actual.Processing, Budget: budget.Processing, Variance: actual.Processing - budget.Processing, VariancePct: calculateVariancePct(actual.Processing, budget.Processing)},
		GA:                  VarianceMetric{Actual: actual.GA, Budget: budget.GA, Variance: actual.GA - budget.GA, VariancePct: calculateVariancePct(actual.GA, budget.GA)},
		TransportShipping:   VarianceMetric{Actual: actual.TransportShipping, Budget: budget.TransportShipping, Variance: actual.TransportShipping - budget.TransportShipping, VariancePct: calculateVariancePct(actual.TransportShipping, budget.TransportShipping)},
		InventoryVariations: VarianceMetric{Actual: actual.InventoryVariations, Budget: budget.InventoryVariations, Variance: actual.InventoryVariations - budget.InventoryVariations, VariancePct: calculateVariancePct(actual.InventoryVariations, budget.InventoryVariations)},
		Total:               VarianceMetric{Actual: actual.Total, Budget: budget.Total, Variance: actual.Total - budget.Total, VariancePct: calculateVariancePct(actual.Total, budget.Total)},
	}
}

// buildCAPEXMonthlyData builds CAPEX monthly data with variances and aggregations
func (uc *detailUseCase) buildCAPEXMonthlyData(
	year int,
	capexActual, capexBudget []*data.CAPEXData,
	monthsFilter map[int]bool,
) ([]CAPEXMonthlyData, map[string]CAPEXTypeData, map[string]CAPEXCategoryData) {
	capexActualByMonth := groupCAPEXByMonth(capexActual)
	capexBudgetByMonth := groupCAPEXByMonth(capexBudget)

	var months []CAPEXMonthlyData
	typeTotals := make(map[string]struct{ Actual, Budget float64 })
	categoryTotals := make(map[string]struct{ Actual, Budget float64 })

	for month := 1; month <= 12; month++ {
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actual := uc.buildCAPEXDetail(capexActualByMonth[month])
		budget := uc.buildCAPEXDetail(capexBudgetByMonth[month])

		// Aggregate by type
		if actual != nil {
			typeTotals["sustaining"] = struct{ Actual, Budget float64 }{typeTotals["sustaining"].Actual + actual.Sustaining, typeTotals["sustaining"].Budget}
			typeTotals["project"] = struct{ Actual, Budget float64 }{typeTotals["project"].Actual + actual.Project, typeTotals["project"].Budget}
			typeTotals["leasing"] = struct{ Actual, Budget float64 }{typeTotals["leasing"].Actual + actual.Leasing, typeTotals["leasing"].Budget}
		}
		if budget != nil {
			typeTotals["sustaining"] = struct{ Actual, Budget float64 }{typeTotals["sustaining"].Actual, typeTotals["sustaining"].Budget + budget.Sustaining}
			typeTotals["project"] = struct{ Actual, Budget float64 }{typeTotals["project"].Actual, typeTotals["project"].Budget + budget.Project}
			typeTotals["leasing"] = struct{ Actual, Budget float64 }{typeTotals["leasing"].Actual, typeTotals["leasing"].Budget + budget.Leasing}
		}

		// Aggregate by category (from raw data)
		for _, capex := range capexActualByMonth[month] {
			categoryTotals[capex.Category] = struct{ Actual, Budget float64 }{categoryTotals[capex.Category].Actual + capex.Amount, categoryTotals[capex.Category].Budget}
		}
		for _, capex := range capexBudgetByMonth[month] {
			categoryTotals[capex.Category] = struct{ Actual, Budget float64 }{categoryTotals[capex.Category].Actual, categoryTotals[capex.Category].Budget + capex.Amount}
		}

		var variance *CAPEXVarianceDetail
		if actual != nil && budget != nil {
			variance = uc.calculateCAPEXVariance(actual, budget)
		}

		months = append(months, CAPEXMonthlyData{
			Month:    monthKey,
			Actual:   actual,
			Budget:   budget,
			Variance: variance,
		})
	}

	// Build type aggregations
	byType := make(map[string]CAPEXTypeData)
	for t, totals := range typeTotals {
		byType[t] = CAPEXTypeData{
			Type:     t,
			Actual:   totals.Actual,
			Budget:   totals.Budget,
			Variance: VarianceMetric{Actual: totals.Actual, Budget: totals.Budget, Variance: totals.Actual - totals.Budget, VariancePct: calculateVariancePct(totals.Actual, totals.Budget)},
		}
	}

	// Build category aggregations
	byCategory := make(map[string]CAPEXCategoryData)
	for cat, totals := range categoryTotals {
		byCategory[cat] = CAPEXCategoryData{
			Category: cat,
			Actual:   totals.Actual,
			Budget:   totals.Budget,
			Variance: VarianceMetric{Actual: totals.Actual, Budget: totals.Budget, Variance: totals.Actual - totals.Budget, VariancePct: calculateVariancePct(totals.Actual, totals.Budget)},
		}
	}

	return months, byType, byCategory
}

// Required CAPEX categories - all must be present in response
var requiredCAPEXCategories = []string{
	// Sustaining Capital (PBR)
	"Pre-Stripping and Capital Developments",
	"Exploration/Mine Geology",
	"Mine Equipment",
	"Mine Infrastructure",
	"Tailings Dams and Leach Pads",
	"Plant Upgrades",
	"Site Infrastructure",
	"Administration Projects",
	"Community Projects",
	"Right-of-Use Asset (IFRS16)",
	// MPPE Additions to Sustaining Capital Reconciliation
	"Total MPPE Additions",
	"Project Capital",
	"Leasing Addition - Project Capital",
	"Other",
	"Sustaining MPPE Additions",
	"Leasing Addition - Sustaining Capital",
	"Sustaining Capital Lease Cash Outflows",
	// Capital Lease
	"IFRS16",
}

// Required CAPEX projects - all must be present in response
var requiredCAPEXProjects = []string{
	"C487EY21001 - CAPEX EXPLORACIONES",
	"C487MY25001",
	"C487MY25002",
	"C487MY25003",
	"C487MY25004",
	"C487MY25005",
	"C487MY25006",
	"C487MY25007",
	"C487MY25008",
	"C487MY25009",
	"C487MY25010",
	"C487PY25001",
	"C487AY25001",
	"C487AY25002",
	"C487AY25003",
	"C487AY24001",
	"C487AY24005",
	"C487AY24003",
	"C48703300",
}

func (uc *detailUseCase) buildCAPEXDetail(capexList []*data.CAPEXData) *CAPEXDetail {
	if len(capexList) == 0 {
		return nil
	}

	var sustaining, project, leasing, accretion float64

	// Initialize maps with all required keys set to 0
	byCategory := make(map[string]float64)
	for _, cat := range requiredCAPEXCategories {
		byCategory[cat] = 0
	}

	byProject := make(map[string]float64)
	for _, proj := range requiredCAPEXProjects {
		byProject[proj] = 0
	}

	for _, capex := range capexList {
		switch capex.Type {
		case "sustaining":
			sustaining += capex.Amount
		case "project":
			project += capex.Amount
		case "leasing":
			leasing += capex.Amount
		}
		// Accretion of Mine Closure Liability - now comes from the field
		accretion += capex.AccretionOfMineClosureLiability

		// Aggregate by category
		if capex.Category != "" {
			byCategory[capex.Category] += capex.Amount
		}

		// Aggregate by project (CAR number + project name)
		projectKey := buildProjectKey(capex.CARNumber, capex.ProjectName)
		if projectKey != "" {
			byProject[projectKey] += capex.Amount
		}
	}

	total := sustaining + project + leasing + accretion

	return &CAPEXDetail{
		Sustaining:                      sustaining,
		Project:                         project,
		Leasing:                         leasing,
		AccretionOfMineClosureLiability: accretion,
		Total:                           total,
		ByCategory:                      byCategory,
		ByProject:                       byProject,
		HasData:                         true,
	}
}

// buildProjectKey creates a project key from CAR number and project name
func buildProjectKey(carNumber, projectName string) string {
	if carNumber == "" {
		return ""
	}
	if projectName == "" || projectName == carNumber {
		return carNumber
	}
	return carNumber + " - " + projectName
}

func (uc *detailUseCase) calculateCAPEXVariance(actual, budget *CAPEXDetail) *CAPEXVarianceDetail {
	return &CAPEXVarianceDetail{
		Sustaining:                      VarianceMetric{Actual: actual.Sustaining, Budget: budget.Sustaining, Variance: actual.Sustaining - budget.Sustaining, VariancePct: calculateVariancePct(actual.Sustaining, budget.Sustaining)},
		Project:                         VarianceMetric{Actual: actual.Project, Budget: budget.Project, Variance: actual.Project - budget.Project, VariancePct: calculateVariancePct(actual.Project, budget.Project)},
		Leasing:                         VarianceMetric{Actual: actual.Leasing, Budget: budget.Leasing, Variance: actual.Leasing - budget.Leasing, VariancePct: calculateVariancePct(actual.Leasing, budget.Leasing)},
		AccretionOfMineClosureLiability: VarianceMetric{Actual: actual.AccretionOfMineClosureLiability, Budget: budget.AccretionOfMineClosureLiability, Variance: actual.AccretionOfMineClosureLiability - budget.AccretionOfMineClosureLiability, VariancePct: calculateVariancePct(actual.AccretionOfMineClosureLiability, budget.AccretionOfMineClosureLiability)},
		Total:                           VarianceMetric{Actual: actual.Total, Budget: budget.Total, Variance: actual.Total - budget.Total, VariancePct: calculateVariancePct(actual.Total, budget.Total)},
	}
}

// buildFinancialMonthlyData builds Financial monthly data with variances
func (uc *detailUseCase) buildFinancialMonthlyData(
	year int,
	financialActual, financialBudget []*data.FinancialData,
	monthsFilter map[int]bool,
) []FinancialMonthlyData {
	financialActualByMonth := groupFinancialByMonth(financialActual)
	financialBudgetByMonth := groupFinancialByMonth(financialBudget)

	var months []FinancialMonthlyData

	for month := 1; month <= 12; month++ {
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actual := uc.buildFinancialDetail(financialActualByMonth[month])
		budget := uc.buildFinancialDetail(financialBudgetByMonth[month])

		var variance *FinancialVariance
		if actual != nil && budget != nil {
			variance = uc.calculateFinancialVariance(actual, budget)
		}

		months = append(months, FinancialMonthlyData{
			Month:    monthKey,
			Actual:   actual,
			Budget:   budget,
			Variance: variance,
		})
	}

	return months
}

func (uc *detailUseCase) buildFinancialDetail(financial *data.FinancialData) *FinancialDetail {
	if financial == nil {
		return nil
	}

	total := financial.ShippingSelling + financial.SalesTaxesRoyalties + financial.OtherAdjustments

	return &FinancialDetail{
		ShippingSelling:     financial.ShippingSelling,
		SalesTaxesRoyalties: financial.SalesTaxesRoyalties,
		OtherAdjustments:    financial.OtherAdjustments,
		Total:               total,
		HasData:             true,
	}
}

func (uc *detailUseCase) calculateFinancialVariance(actual, budget *FinancialDetail) *FinancialVariance {
	return &FinancialVariance{
		ShippingSelling:     VarianceMetric{Actual: actual.ShippingSelling, Budget: budget.ShippingSelling, Variance: actual.ShippingSelling - budget.ShippingSelling, VariancePct: calculateVariancePct(actual.ShippingSelling, budget.ShippingSelling)},
		SalesTaxesRoyalties: VarianceMetric{Actual: actual.SalesTaxesRoyalties, Budget: budget.SalesTaxesRoyalties, Variance: actual.SalesTaxesRoyalties - budget.SalesTaxesRoyalties, VariancePct: calculateVariancePct(actual.SalesTaxesRoyalties, budget.SalesTaxesRoyalties)},
		OtherAdjustments:    VarianceMetric{Actual: actual.OtherAdjustments, Budget: budget.OtherAdjustments, Variance: actual.OtherAdjustments - budget.OtherAdjustments, VariancePct: calculateVariancePct(actual.OtherAdjustments, budget.OtherAdjustments)},
		Total:               VarianceMetric{Actual: actual.Total, Budget: budget.Total, Variance: actual.Total - budget.Total, VariancePct: calculateVariancePct(actual.Total, budget.Total)},
	}
}

// buildProductionMonthlyData builds Production monthly data with variances
func (uc *detailUseCase) buildProductionMonthlyData(
	year int,
	pbrActual, pbrBudget []*data.PBRData,
	productionActual, productionBudget []*data.ProductionData,
	mineralMap map[int]struct{ Code, Name string },
	monthsFilter map[int]bool,
) ([]ProductionMonthlyData, map[string]ProductionMineralData) {
	pbrActualByMonth := groupPBRByMonth(pbrActual)
	pbrBudgetByMonth := groupPBRByMonth(pbrBudget)
	productionActualByMonth := groupProductionByMonth(productionActual)
	productionBudgetByMonth := groupProductionByMonth(productionBudget)

	var months []ProductionMonthlyData
	mineralTotals := make(map[string]struct {
		MineralName string
		Unit        string
		Actual      float64
		Budget      float64
	})

	for month := 1; month <= 12; month++ {
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actual := uc.buildProductionDetail(pbrActualByMonth[month], productionActualByMonth[month], mineralMap)
		budget := uc.buildProductionDetail(pbrBudgetByMonth[month], productionBudgetByMonth[month], mineralMap)

		// Aggregate by mineral
		if actual != nil {
			for code, qty := range actual.ByMineral {
				if _, exists := mineralTotals[code]; !exists {
					// Find mineral name from map (reverse lookup)
					var mineralName string
					for _, m := range mineralMap {
						if m.Code == code {
							mineralName = m.Name
							break
						}
					}
					if mineralName == "" {
						mineralName = code
					}
					mineralTotals[code] = struct {
						MineralName string
						Unit        string
						Actual      float64
						Budget      float64
					}{MineralName: mineralName, Unit: "", Actual: 0, Budget: 0}
				}
				totals := mineralTotals[code]
				totals.Actual += qty
				mineralTotals[code] = totals
			}
		}
		if budget != nil {
			for code, qty := range budget.ByMineral {
				if _, exists := mineralTotals[code]; !exists {
					// Find mineral name from map (reverse lookup)
					var mineralName string
					for _, m := range mineralMap {
						if m.Code == code {
							mineralName = m.Name
							break
						}
					}
					if mineralName == "" {
						mineralName = code
					}
					mineralTotals[code] = struct {
						MineralName string
						Unit        string
						Actual      float64
						Budget      float64
					}{MineralName: mineralName, Unit: "", Actual: 0, Budget: 0}
				}
				totals := mineralTotals[code]
				totals.Budget += qty
				mineralTotals[code] = totals
			}
		}

		var variance *ProductionVarianceDetail
		if actual != nil && budget != nil {
			variance = uc.calculateProductionVariance(actual, budget)
		}

		months = append(months, ProductionMonthlyData{
			Month:    monthKey,
			Actual:   actual,
			Budget:   budget,
			Variance: variance,
		})
	}

	// Build mineral aggregations
	byMineral := make(map[string]ProductionMineralData)
	for code, totals := range mineralTotals {
		byMineral[code] = ProductionMineralData{
			MineralCode: code,
			MineralName: totals.MineralName,
			Unit:        totals.Unit,
			Actual:      totals.Actual,
			Budget:      totals.Budget,
			Variance:    VarianceMetric{Actual: totals.Actual, Budget: totals.Budget, Variance: totals.Actual - totals.Budget, VariancePct: calculateVariancePct(totals.Actual, totals.Budget)},
		}
	}

	return months, byMineral
}

func (uc *detailUseCase) buildProductionDetail(pbr *data.PBRData, productionList []*data.ProductionData, mineralMap map[int]struct{ Code, Name string }) *ProductionDetail {
	byMineral := make(map[string]float64)

	// Add Silver and Gold from PBR
	if pbr != nil {
		silverOz := pbr.FeedGradeSilverGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateSilverPct / 100) / 31.1035
		goldOz := pbr.FeedGradeGoldGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateGoldPct / 100) / 31.1035
		byMineral["AG"] = silverOz
		byMineral["AU"] = goldOz
	}

	// Add other minerals from ProductionData
	for _, prod := range productionList {
		if mineral, exists := mineralMap[prod.MineralID]; exists {
			byMineral[mineral.Code] += prod.Quantity
		}
	}

	hasData := pbr != nil || len(productionList) > 0

	var silverOz, goldOz float64
	if pbr != nil {
		silverOz = pbr.FeedGradeSilverGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateSilverPct / 100) / 31.1035
		goldOz = pbr.FeedGradeGoldGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateGoldPct / 100) / 31.1035
	}

	return &ProductionDetail{
		TotalProductionSilverOz: silverOz,
		TotalProductionGoldOz:   goldOz,
		ByMineral:               byMineral,
		HasData:                 hasData,
	}
}

func (uc *detailUseCase) calculateProductionVariance(actual, budget *ProductionDetail) *ProductionVarianceDetail {
	return &ProductionVarianceDetail{
		TotalProductionSilverOz: VarianceMetric{Actual: actual.TotalProductionSilverOz, Budget: budget.TotalProductionSilverOz, Variance: actual.TotalProductionSilverOz - budget.TotalProductionSilverOz, VariancePct: calculateVariancePct(actual.TotalProductionSilverOz, budget.TotalProductionSilverOz)},
		TotalProductionGoldOz:   VarianceMetric{Actual: actual.TotalProductionGoldOz, Budget: budget.TotalProductionGoldOz, Variance: actual.TotalProductionGoldOz - budget.TotalProductionGoldOz, VariancePct: calculateVariancePct(actual.TotalProductionGoldOz, budget.TotalProductionGoldOz)},
	}
}

// buildRevenueMonthlyData builds Revenue monthly data with variances
func (uc *detailUseCase) buildRevenueMonthlyData(
	year int,
	revenueActual, revenueBudget []*data.RevenueData,
	mineralMap map[int]struct{ Code, Name string },
	monthsFilter map[int]bool,
) ([]RevenueMonthlyData, map[string]RevenueMineralData) {
	revenueActualByMonth := groupRevenueByMonth(revenueActual)
	revenueBudgetByMonth := groupRevenueByMonth(revenueBudget)

	var months []RevenueMonthlyData
	mineralTotals := make(map[string]struct {
		MineralName string
		Currency    string
		Actual      float64
		Budget      float64
	})

	for month := 1; month <= 12; month++ {
		if monthsFilter != nil && !monthsFilter[month] {
			continue
		}

		monthKey := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

		actual := uc.buildRevenueDetail(revenueActualByMonth[month], mineralMap)
		budget := uc.buildRevenueDetail(revenueBudgetByMonth[month], mineralMap)

		// Aggregate by mineral
		if actual != nil {
			for code, detail := range actual.ByMineral {
				if _, exists := mineralTotals[code]; !exists {
					mineralTotals[code] = struct {
						MineralName string
						Currency    string
						Actual      float64
						Budget      float64
					}{MineralName: detail.MineralName, Currency: detail.Currency, Actual: 0, Budget: 0}
				}
				totals := mineralTotals[code]
				totals.Actual += detail.Revenue
				mineralTotals[code] = totals
			}
		}
		if budget != nil {
			for code, detail := range budget.ByMineral {
				if _, exists := mineralTotals[code]; !exists {
					mineralTotals[code] = struct {
						MineralName string
						Currency    string
						Actual      float64
						Budget      float64
					}{MineralName: detail.MineralName, Currency: detail.Currency, Actual: 0, Budget: 0}
				}
				totals := mineralTotals[code]
				totals.Budget += detail.Revenue
				mineralTotals[code] = totals
			}
		}

		var variance *RevenueVariance
		if actual != nil && budget != nil {
			variance = uc.calculateRevenueVariance(actual, budget)
		}

		months = append(months, RevenueMonthlyData{
			Month:    monthKey,
			Actual:   actual,
			Budget:   budget,
			Variance: variance,
		})
	}

	// Build mineral aggregations
	byMineral := make(map[string]RevenueMineralData)
	for code, totals := range mineralTotals {
		byMineral[code] = RevenueMineralData{
			MineralCode: code,
			MineralName: totals.MineralName,
			Currency:    totals.Currency,
			Actual:      totals.Actual,
			Budget:      totals.Budget,
			Variance:    VarianceMetric{Actual: totals.Actual, Budget: totals.Budget, Variance: totals.Actual - totals.Budget, VariancePct: calculateVariancePct(totals.Actual, totals.Budget)},
		}
	}

	return months, byMineral
}

func (uc *detailUseCase) buildRevenueDetail(revenueList []*data.RevenueData, mineralMap map[int]struct{ Code, Name string }) *RevenueDetail {
	if len(revenueList) == 0 {
		return nil
	}

	byMineral := make(map[string]RevenueMineralDetail)
	var totalRevenue, totalQuantity float64

	for _, rev := range revenueList {
		var code, name string
		if mineral, exists := mineralMap[rev.MineralID]; exists {
			code = mineral.Code
			name = mineral.Name
		} else {
			code = "UNKNOWN"
			name = "Unknown Mineral"
		}

		revenue := rev.QuantitySold * rev.UnitPrice

		if existing, exists := byMineral[code]; exists {
			existing.QuantitySold += rev.QuantitySold
			existing.Revenue += revenue
			byMineral[code] = existing
		} else {
			byMineral[code] = RevenueMineralDetail{
				MineralCode:  code,
				MineralName:  name,
				QuantitySold: rev.QuantitySold,
				UnitPrice:    rev.UnitPrice,
				Revenue:      revenue,
				Currency:     rev.Currency,
			}
		}

		totalRevenue += revenue
		totalQuantity += rev.QuantitySold
	}

	var avgUnitPrice float64
	if totalQuantity > 0 {
		avgUnitPrice = totalRevenue / totalQuantity
	}

	return &RevenueDetail{
		ByMineral:         byMineral,
		TotalRevenue:      totalRevenue,
		TotalQuantitySold: totalQuantity,
		AverageUnitPrice:  avgUnitPrice,
		HasData:           true,
	}
}

func (uc *detailUseCase) calculateRevenueVariance(actual, budget *RevenueDetail) *RevenueVariance {
	return &RevenueVariance{
		TotalRevenue:      VarianceMetric{Actual: actual.TotalRevenue, Budget: budget.TotalRevenue, Variance: actual.TotalRevenue - budget.TotalRevenue, VariancePct: calculateVariancePct(actual.TotalRevenue, budget.TotalRevenue)},
		TotalQuantitySold: VarianceMetric{Actual: actual.TotalQuantitySold, Budget: budget.TotalQuantitySold, Variance: actual.TotalQuantitySold - budget.TotalQuantitySold, VariancePct: calculateVariancePct(actual.TotalQuantitySold, budget.TotalQuantitySold)},
		AverageUnitPrice:  VarianceMetric{Actual: actual.AverageUnitPrice, Budget: budget.AverageUnitPrice, Variance: actual.AverageUnitPrice - budget.AverageUnitPrice, VariancePct: calculateVariancePct(actual.AverageUnitPrice, budget.AverageUnitPrice)},
	}
}

// Helper functions to group data by month (reuse from usecase.go)
func groupProductionByMonth(records []*data.ProductionData) map[int][]*data.ProductionData {
	grouped := make(map[int][]*data.ProductionData)
	for _, r := range records {
		month := int(r.Date.Month())
		grouped[month] = append(grouped[month], r)
	}
	return grouped
}

func groupRevenueByMonth(records []*data.RevenueData) map[int][]*data.RevenueData {
	grouped := make(map[int][]*data.RevenueData)
	for _, r := range records {
		month := int(r.Date.Month())
		grouped[month] = append(grouped[month], r)
	}
	return grouped
}
