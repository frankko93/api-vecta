package reports

import (
	"time"

	"github.com/gmhafiz/go8/internal/domain/data"
)

// Test constants - SINGLE SOURCE OF TRUTH
const (
	testCompanyID = int64(1)
	testUserID    = int64(1)
)

// Data factories - Use these in ALL calculator tests
// These values match Cerro Moro January 2025 real data

// newTestPBRData creates standard PBR test data
// Represents real data from Cerro Moro January 2025
func newTestPBRData() *data.PBRData {
	return &data.PBRData{
		CompanyID:             testCompanyID,
		Date:                  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		OreMinedT:             24859,
		WasteMinedT:           262591,
		DevelopmentsM:         598,
		TotalTonnesProcessed:  35951,
		FeedGradeSilverGpt:    209.79,
		FeedGradeGoldGpt:      7.35,
		RecoveryRateSilverPct: 94.01,
		RecoveryRateGoldPct:   95.36,
		DataType:              "actual",
		CreatedBy:             testUserID,
	}
}

// newTestDoreData creates standard Dore test data
func newTestDoreData() *data.DoreData {
	return &data.DoreData{
		CompanyID:            testCompanyID,
		Date:                 time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		DoreProducedOz:       236064,
		SilverGradePct:       85.5,
		GoldGradePct:         14.5,
		PBRPriceSilver:       24.50,
		PBRPriceGold:         2000,
		RealizedPriceSilver:  24.30,
		RealizedPriceGold:    1985,
		SilverAdjustmentOz:   10,
		GoldAdjustmentOz:     5,
		AgDeductionsPct:      2.5,
		AuDeductionsPct:      1.5,
		TreatmentCharge:      5000,
		RefiningDeductionsAu: 1200,
		DataType:             "actual",
		CreatedBy:            testUserID,
	}
}

// newTestFinancialData creates standard financial test data
func newTestFinancialData() *data.FinancialData {
	return &data.FinancialData{
		CompanyID:           testCompanyID,
		Date:                time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		ShippingSelling:     -202,
		SalesTaxesRoyalties: 465867,
		OtherAdjustments:    0,
		Currency:            "USD",
		DataType:            "actual",
		CreatedBy:           testUserID,
	}
}

// newTestOPEXList creates standard OPEX test data
// Matches Cerro Moro January 2025 costs breakdown
func newTestOPEXList() []*data.OPEXData {
	baseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	
	return []*data.OPEXData{
		{
			CompanyID:   testCompanyID,
			Date:        baseDate,
			CostCenter:  "Mine",
			Subcategory: "Drilling",
			ExpenseType: "Labour",
			Amount:      8537997,
			Currency:    "USD",
			DataType:    "actual",
			CreatedBy:   testUserID,
		},
		{
			CompanyID:   testCompanyID,
			Date:        baseDate,
			CostCenter:  "Processing",
			Subcategory: "CO General Operating",
			ExpenseType: "Labour",
			Amount:      3613678,
			Currency:    "USD",
			DataType:    "actual",
			CreatedBy:   testUserID,
		},
		{
			CompanyID:   testCompanyID,
			Date:        baseDate,
			CostCenter:  "G&A",
			Subcategory: "General Administration",
			ExpenseType: "Third Party",
			Amount:      5471220,
			Currency:    "USD",
			DataType:    "actual",
			CreatedBy:   testUserID,
		},
		{
			CompanyID:   testCompanyID,
			Date:        baseDate,
			CostCenter:  "Mine",
			Subcategory: "Inventory Variations",
			ExpenseType: "Other",
			Amount:      1740162,
			Currency:    "USD",
			DataType:    "actual",
			CreatedBy:   testUserID,
		},
	}
}

// newTestCAPEXList creates standard CAPEX test data
// Total sustaining: 711,052 (matches Cerro Moro)
func newTestCAPEXList() []*data.CAPEXData {
	baseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	
	return []*data.CAPEXData{
		{
			CompanyID:   testCompanyID,
			Date:        baseDate,
			Category:    "Mine Equipment",
			CARNumber:   "C487MY25001",
			ProjectName: "Equipment Purchase",
			Type:        "sustaining",
			Amount:      500000,
			Currency:    "USD",
			DataType:    "actual",
			CreatedBy:   testUserID,
		},
		{
			CompanyID:   testCompanyID,
			Date:        baseDate,
			Category:    "Mine Infrastructure",
			ProjectName: "Infrastructure Upgrade",
			Type:        "sustaining",
			Amount:      211052,
			Currency:    "USD",
			DataType:    "actual",
			CreatedBy:   testUserID,
		},
		{
			CompanyID:   testCompanyID,
			Date:        baseDate,
			Category:    "Plant Upgrades",
			ProjectName: "Plant Expansion",
			Type:        "project",
			Amount:      350000,
			Currency:    "USD",
			DataType:    "actual",
			CreatedBy:   testUserID,
		},
		{
			CompanyID:   testCompanyID,
			Date:        baseDate,
			Category:    "Equipment Lease",
			ProjectName: "Equipment Leasing",
			Type:        "leasing",
			Amount:      100000,
			Currency:    "USD",
			DataType:    "actual",
			CreatedBy:   testUserID,
		},
	}
}

// Expected values from Cerro Moro Summary - Use these for validation
const (
	expectedTotalProductionSilverOz = 227957.0
	expectedTotalProductionGoldOz   = 8106.0
	expectedProductionBasedCosts    = 19363057.0
	expectedSustainingCAPEX         = 711052.0
	expectedNSRDore                 = 27919108.0
	expectedNetSmelterReturn        = 27453443.0
)

