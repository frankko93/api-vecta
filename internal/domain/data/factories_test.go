package data

// Test constants - SINGLE SOURCE OF TRUTH for all data tests
const (
	testCompanyID   = int64(1)
	testUserID      = int64(1)
	testVersion     = 1
	testDescription = "test data"
)

// CSV builders - Use these in ALL CSV parsing tests to maintain consistency

func buildProductionCSV(rows []string) []byte {
	csv := "date,mineral_code,quantity,unit\n"
	for _, row := range rows {
		csv += row + "\n"
	}
	return []byte(csv)
}

func buildPBRCSV(rows []string) []byte {
	csv := "date,ore_mined_t,waste_mined_t,developments_m,total_tonnes_processed,feed_grade_silver_gpt,feed_grade_gold_gpt,recovery_rate_silver_pct,recovery_rate_gold_pct\n"
	for _, row := range rows {
		csv += row + "\n"
	}
	return []byte(csv)
}

func buildOPEXCSV(rows []string) []byte {
	csv := "date,cost_center,subcategory,expense_type,amount,currency\n"
	for _, row := range rows {
		csv += row + "\n"
	}
	return []byte(csv)
}

func buildFinancialCSV(rows []string) []byte {
	csv := "date,shipping_selling,sales_taxes_royalties,other_adjustments\n"
	for _, row := range rows {
		csv += row + "\n"
	}
	return []byte(csv)
}

// Test mineral map - SINGLE SOURCE OF TRUTH
// If you add/remove minerals, update this and all tests will use the new data
func getTestMineralMap() map[string]int {
	return map[string]int{
		"AU": 1,
		"AG": 2,
		"CU": 3,
		"ZN": 4,
		"PB": 5,
		"LI": 6,
		"FE": 7,
	}
}

// Valid test data rows - REFERENCE DATA
// Use these constants to ensure consistency across tests
const (
	validProductionRow = "2024-01-15,AU,150.5,kilograms"
	validPBRRow        = "2024-01-15,24859,262591,598,35951,209.79,7.35,94.01,95.36"
	validOPEXRow       = "2024-01-15,Mine,Drilling,Labour,50000,USD"
	validFinancialRow  = "2024-01-15,-202,465867,0"
)
