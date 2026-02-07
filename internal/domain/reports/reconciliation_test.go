package reports

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gmhafiz/go8/internal/domain/data"
)

// ReferenceSummary represents parsed values from the reference Summary.csv
type ReferenceSummary struct {
	Month  int // 1-12
	Values map[string]ReferenceValue
}

// ReferenceValue represents a single metric value from reference Summary
type ReferenceValue struct {
	Actual      float64
	Budget      float64
	Variance    float64
	VariancePct float64
	YTDActual   float64
	YTDBudget   float64
	YTDVariance float64
	YTDVariancePct float64
}

// MetricMapping maps Summary.csv row labels to DataSet field paths
var metricMapping = map[string]struct {
	Category string
	Field    string
}{
	"Ore Mined (t)":                          {"mining", "ore_mined_t"},
	"Waste Mined (t)":                        {"mining", "waste_mined_t"},
	"Developments (m)":                       {"mining", "developments_m"},
	"Total Tonnes Processed":                  {"processing", "total_tonnes_processed"},
	"Feed Grade - Silver (g/t)":              {"processing", "feed_grade_silver_gpt"},
	"Feed Grade - Gold (g/t)":                {"processing", "feed_grade_gold_gpt"},
	"Recovery Rate - Silver (%)":             {"processing", "recovery_rate_silver_pct"},
	"Recovery Rate - Gold (%)":               {"processing", "recovery_rate_gold_pct"},
	"Total Production - Silver (oz)":         {"production", "total_production_silver_oz"},
	"Total Production - Gold (oz)":           {"production", "total_production_gold_oz"},
	"Payable Metal in Dore - Silver (oz)":   {"production", "payable_silver_oz"},
	"Payable Metal in Dore - Gold (oz)":       {"production", "payable_gold_oz"},
	"NSR per tonne":                           {"nsr", "nsr_per_tonne"},
	"Total cost per tonne":                    {"nsr", "total_cost_per_tonne"},
	"Margin per Tonne":                        {"nsr", "margin_per_tonne"},
	"Net Smelter Return - Dore":               {"nsr", "nsr_dore"},
	"Shipping & Selling":                     {"nsr", "shipping_selling"},
	"Sales Taxes & Royalties":                 {"nsr", "sales_taxes_royalties"},
	"Net Smelter Return":                      {"nsr", "net_smelter_return"},
	"Costs - Mine":                            {"costs", "mine"},
	"Costs - Processing":                      {"costs", "processing"},
	"Costs - G&A":                             {"costs", "ga"},
	"Transport & Shipping":                    {"costs", "transport_shipping"},
	"Inventory Variations":                    {"costs", "inventory_variations"},
	"Production based Costs":                   {"costs", "production_based_costs"},
	"Production based Margin":                 {"costs", "production_based_margin"},
	"AISC Sustaining Capital":                 {"capex", "sustaining"},
	"PBR Net Cash flow":                       {"capex", "pbr_net_cash_flow"},
	"Cash Cost per Payable Ounce - Silver":   {"cash_cost", "cash_cost_per_oz_silver"},
	"AISC per Payable Ounce - Silver":        {"cash_cost", "aisc_per_oz_silver"},
}

// parseReferenceSummary parses the reference Summary.csv file
func parseReferenceSummary(filePath string, month int) (*ReferenceSummary, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open reference file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 4 {
		return nil, fmt.Errorf("invalid CSV format: expected at least 4 rows")
	}

	// Row 0: Header metadata (ignore)
	// Row 1: "Actual,2025 Budget,,,Actual,2025 Budget,,"
	// Row 2: "Jan,Jan,Fav (Unf),% Variance,Jan_YTD,Jan_YTD,Fav (Unf),% Variance"
	// Row 3+: Data rows

	summary := &ReferenceSummary{
		Month:  month,
		Values: make(map[string]ReferenceValue),
	}

	// Find the column indices for the requested month
	// Columns: [Label, Actual, Budget, Variance, VariancePct, YTD Actual, YTD Budget, YTD Variance, YTD VariancePct]
	actualCol := 1
	budgetCol := 2
	varianceCol := 3
	variancePctCol := 4
	ytdActualCol := 5
	ytdBudgetCol := 6
	ytdVarianceCol := 7
	ytdVariancePctCol := 8

	// Parse data rows (starting from row 4, index 3)
	for i := 3; i < len(records); i++ {
		row := records[i]
		if len(row) < 9 {
			continue
		}

		label := strings.TrimSpace(row[0])
		if label == "" {
			continue
		}

		// Check if this metric is in our mapping
		if _, exists := metricMapping[label]; !exists {
			continue
		}

		// Parse values using robust parsing
		actual, _ := parseReferenceValue(row[actualCol])
		budget, _ := parseReferenceValue(row[budgetCol])
		variance, _ := parseReferenceValue(row[varianceCol])
		variancePct, _ := parseReferenceValue(row[variancePctCol])
		ytdActual, _ := parseReferenceValue(row[ytdActualCol])
		ytdBudget, _ := parseReferenceValue(row[ytdBudgetCol])
		ytdVariance, _ := parseReferenceValue(row[ytdVarianceCol])
		ytdVariancePct, _ := parseReferenceValue(row[ytdVariancePctCol])

		summary.Values[label] = ReferenceValue{
			Actual:         actual,
			Budget:         budget,
			Variance:       variance,
			VariancePct:    variancePct,
			YTDActual:      ytdActual,
			YTDBudget:      ytdBudget,
			YTDVariance:    ytdVariance,
			YTDVariancePct: ytdVariancePct,
		}
	}

	return summary, nil
}

// parseReferenceValue parses a value from Summary.csv (handles formatting)
func parseReferenceValue(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "-" || value == "$ -" || value == "$ -   " {
		return 0, nil
	}

	// Remove currency symbols
	value = strings.TrimPrefix(value, "$")
	value = strings.TrimSpace(value)

	// Remove percentage sign
	value = strings.TrimSuffix(value, "%")
	value = strings.TrimSpace(value)

	// Handle parentheses for negatives
	isNegative := false
	if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
		isNegative = true
		value = strings.TrimPrefix(value, "(")
		value = strings.TrimSuffix(value, ")")
		value = strings.TrimSpace(value)
	}

	// Remove commas
	value = strings.ReplaceAll(value, ",", "")

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", value)
	}

	if isNegative {
		f = -f
	}

	return f, nil
}

// getValueFromDataSet extracts a value from DataSet using category and field name
func getValueFromDataSet(ds *DataSet, category, field string) float64 {
	switch category {
	case "mining":
		switch field {
		case "ore_mined_t":
			return ds.Mining.OreMinedT
		case "waste_mined_t":
			return ds.Mining.WasteMinedT
		case "developments_m":
			return ds.Mining.DevelopmentsM
		}
	case "processing":
		switch field {
		case "total_tonnes_processed":
			return ds.Processing.TotalTonnesProcessed
		case "feed_grade_silver_gpt":
			return ds.Processing.FeedGradeSilverGpt
		case "feed_grade_gold_gpt":
			return ds.Processing.FeedGradeGoldGpt
		case "recovery_rate_silver_pct":
			return ds.Processing.RecoveryRateSilverPct
		case "recovery_rate_gold_pct":
			return ds.Processing.RecoveryRateGoldPct
		}
	case "production":
		switch field {
		case "total_production_silver_oz":
			return ds.Production.TotalProductionSilverOz
		case "total_production_gold_oz":
			return ds.Production.TotalProductionGoldOz
		case "payable_silver_oz":
			return ds.Production.PayableSilverOz
		case "payable_gold_oz":
			return ds.Production.PayableGoldOz
		}
	case "nsr":
		switch field {
		case "nsr_per_tonne":
			return ds.NSR.NSRPerTonne
		case "total_cost_per_tonne":
			return ds.NSR.TotalCostPerTonne
		case "margin_per_tonne":
			return ds.NSR.MarginPerTonne
		case "nsr_dore":
			return ds.NSR.NSRDore
		case "shipping_selling":
			return ds.NSR.ShippingSelling
		case "sales_taxes_royalties":
			return ds.NSR.SalesTaxesRoyalties
		case "net_smelter_return":
			return ds.NSR.NetSmelterReturn
		}
	case "costs":
		switch field {
		case "mine":
			return ds.Costs.Mine
		case "processing":
			return ds.Costs.Processing
		case "ga":
			return ds.Costs.GA
		case "transport_shipping":
			return ds.Costs.TransportShipping
		case "inventory_variations":
			return ds.Costs.InventoryVariations
		case "production_based_costs":
			return ds.Costs.ProductionBasedCosts
		case "production_based_margin":
			return ds.Costs.ProductionBasedMargin
		}
	case "capex":
		switch field {
		case "sustaining":
			return ds.CAPEX.Sustaining
		case "pbr_net_cash_flow":
			return ds.CAPEX.PBRNetCashFlow
		}
	case "cash_cost":
		switch field {
		case "cash_cost_per_oz_silver":
			return ds.CashCost.CashCostPerOzSilver
		case "aisc_per_oz_silver":
			return ds.CashCost.AISCPerOzSilver
		}
	}
	return 0
}

// ReconciliationResult represents the result of comparing API vs Reference
type ReconciliationResult struct {
	Matches    []MetricMatch
	Mismatches []MetricMismatch
	Summary    ReconciliationSummary
}

// MetricMatch represents a metric that matches within tolerance
type MetricMatch struct {
	Category      string
	MetricName    string
	ActualValue   float64
	ExpectedValue float64
	Difference    float64
	DifferencePct float64
}

// MetricMismatch represents a metric that doesn't match
type MetricMismatch struct {
	Category       string
	MetricName     string
	ActualValue    float64
	ExpectedValue  float64
	Difference     float64
	DifferencePct  float64
	DependencyChain []string
}

// ReconciliationSummary provides overall statistics
type ReconciliationSummary struct {
	TotalMetrics   int
	Matches       int
	Mismatches    int
	MatchRate     float64
	MaxDifference float64
	MaxDifferencePct float64
}

// Reconcile compares API-calculated Summary with reference Summary.csv
func Reconcile(apiActual, apiBudget *DataSet, reference *ReferenceSummary, tolerance float64) *ReconciliationResult {
	result := &ReconciliationResult{
		Matches:    []MetricMatch{},
		Mismatches: []MetricMismatch{},
	}

	for label, refValue := range reference.Values {
		mapping, exists := metricMapping[label]
		if !exists {
			continue
		}

		// Get API values
		apiActualVal := getValueFromDataSet(apiActual, mapping.Category, mapping.Field)
		apiBudgetVal := getValueFromDataSet(apiBudget, mapping.Category, mapping.Field)

		// Compare Actual
		actualDiff := apiActualVal - refValue.Actual
		actualDiffPct := 0.0
		if refValue.Actual != 0 {
			actualDiffPct = (actualDiff / refValue.Actual) * 100
		}

		// Compare Budget
		budgetDiff := apiBudgetVal - refValue.Budget
		budgetDiffPct := 0.0
		if refValue.Budget != 0 {
			budgetDiffPct = (budgetDiff / refValue.Budget) * 100
		}

		// Check if within tolerance (use absolute difference or percentage, whichever is larger)
		actualWithinTolerance := abs(actualDiff) <= tolerance || abs(actualDiffPct) <= tolerance
		budgetWithinTolerance := abs(budgetDiff) <= tolerance || abs(budgetDiffPct) <= tolerance

		if actualWithinTolerance && budgetWithinTolerance {
			result.Matches = append(result.Matches, MetricMatch{
				Category:      mapping.Category,
				MetricName:    label,
				ActualValue:   apiActualVal,
				ExpectedValue: refValue.Actual,
				Difference:    actualDiff,
				DifferencePct: actualDiffPct,
			})
		} else {
			result.Mismatches = append(result.Mismatches, MetricMismatch{
				Category:       mapping.Category,
				MetricName:     label,
				ActualValue:    apiActualVal,
				ExpectedValue:  refValue.Actual,
				Difference:     actualDiff,
				DifferencePct:  actualDiffPct,
				DependencyChain: []string{mapping.Category, mapping.Field},
			})
		}
	}

	// Calculate summary
	result.Summary.TotalMetrics = len(result.Matches) + len(result.Mismatches)
	result.Summary.Matches = len(result.Matches)
	result.Summary.Mismatches = len(result.Mismatches)
	if result.Summary.TotalMetrics > 0 {
		result.Summary.MatchRate = float64(result.Summary.Matches) / float64(result.Summary.TotalMetrics) * 100
	}

	// Find max differences
	for _, m := range result.Mismatches {
		if abs(m.Difference) > result.Summary.MaxDifference {
			result.Summary.MaxDifference = abs(m.Difference)
		}
		if abs(m.DifferencePct) > result.Summary.MaxDifferencePct {
			result.Summary.MaxDifferencePct = abs(m.DifferencePct)
		}
	}

	return result
}

// abs returns absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GenerateDiffReport generates a human-readable diff report
func GenerateDiffReport(result *ReconciliationResult) string {
	var report strings.Builder

	report.WriteString("=== RECONCILIATION DIFF REPORT ===\n\n")
	report.WriteString(fmt.Sprintf("Summary:\n"))
	report.WriteString(fmt.Sprintf("  Total Metrics: %d\n", result.Summary.TotalMetrics))
	report.WriteString(fmt.Sprintf("  Matches: %d (%.2f%%)\n", result.Summary.Matches, result.Summary.MatchRate))
	report.WriteString(fmt.Sprintf("  Mismatches: %d\n", result.Summary.Mismatches))
	report.WriteString(fmt.Sprintf("  Max Difference: %.2f\n", result.Summary.MaxDifference))
	report.WriteString(fmt.Sprintf("  Max Difference %%: %.2f%%\n\n", result.Summary.MaxDifferencePct))

	if len(result.Mismatches) > 0 {
		report.WriteString("MISMATCHES:\n")
		report.WriteString("-----------\n")
		for _, m := range result.Mismatches {
			report.WriteString(fmt.Sprintf("\n%s [%s]\n", m.MetricName, m.Category))
			report.WriteString(fmt.Sprintf("  Expected: %.2f\n", m.ExpectedValue))
			report.WriteString(fmt.Sprintf("  Actual:   %.2f\n", m.ActualValue))
			report.WriteString(fmt.Sprintf("  Diff:     %.2f (%.2f%%)\n", m.Difference, m.DifferencePct))
			if len(m.DependencyChain) > 0 {
				report.WriteString(fmt.Sprintf("  Dependencies: %s\n", strings.Join(m.DependencyChain, " -> ")))
			}
		}
		report.WriteString("\n")
	}

	if len(result.Matches) > 0 {
		report.WriteString(fmt.Sprintf("\nMATCHES: %d metrics within tolerance\n", len(result.Matches)))
	}

	return report.String()
}

// TestReconciliation is the main test function
// It imports sample CSVs, calculates Summary, and compares with reference
// Set RECONCILIATION_REF_PATH environment variable to specify custom path
func TestReconciliation(t *testing.T) {
	// Get reference path from environment or use default
	referencePath := os.Getenv("RECONCILIATION_REF_PATH")
	if referencePath == "" {
		// Default path
		referencePath = filepath.Join("..", "..", "..", "..", "Downloads", "2025_01 PBR_Report-CerroMoro_ (PAS_Corp).xlsx - Summary.csv")
	}
	
	// Try to find the file
	if _, err := os.Stat(referencePath); os.IsNotExist(err) {
		t.Skipf("Reference Summary.csv not found at %s. Set RECONCILIATION_REF_PATH env var to specify path. Skipping reconciliation test.", referencePath)
		return
	}
	
	t.Logf("Using reference file: %s", referencePath)

	// Parse reference Summary for January
	reference, err := parseReferenceSummary(referencePath, 1)
	if err != nil {
		t.Fatalf("Failed to parse reference Summary: %v", err)
	}

	// Create calculator
	calc := NewCalculator()

	// Create test data matching the reference (from actual_2025_ene_*.csv files)
	// Note: These values should match what's in the reference Summary
	pbrActual := &data.PBRData{
		OreMinedT:             24859,
		WasteMinedT:           262591,
		DevelopmentsM:         598,
		TotalTonnesProcessed:  35951,
		FeedGradeSilverGpt:    209.79,
		FeedGradeGoldGpt:      7.35,
		RecoveryRateSilverPct: 94.01,
		RecoveryRateGoldPct:   95.36,
	}

	// Calculate production from PBR (using CalculateDataSet to get production)
	dsActual := calc.CalculateDataSet(pbrActual, nil, nil, nil, nil)
	productionActual := dsActual.Production

	// Create Dore data (values from reference Summary calculations)
	// Note: These need to match the reference values
	doreActual := &data.DoreData{
		DoreProducedOz:       productionActual.DoreProductionOz,
		SilverGradePct:       (productionActual.TotalProductionSilverOz / productionActual.DoreProductionOz) * 100,
		GoldGradePct:         (productionActual.TotalProductionGoldOz / productionActual.DoreProductionOz) * 100,
		PBRPriceSilver:       24.30,
		PBRPriceGold:         1985,
		RealizedPriceSilver:  24.20,
		RealizedPriceGold:    1980,
		SilverAdjustmentOz:   -50,
		GoldAdjustmentOz:     5,
		AgDeductionsPct:     2.5,
		AuDeductionsPct:      1.8,
		TreatmentCharge:      4800,
		RefiningDeductionsAu: 1200,
	}

	financialActual := &data.FinancialData{
		ShippingSelling:     -202,
		SalesTaxesRoyalties: 465867,
	}

	opexActual := []*data.OPEXData{
		{CostCenter: "Mine", Amount: 700000 + 600000 + 450000}, // Sum of Mine costs
		{CostCenter: "Processing", Amount: 350000},
		{CostCenter: "G&A", Amount: 400000},
		{Subcategory: "Inventory Variations", Amount: 1740162},
	}

	capexActual := []*data.CAPEXData{
		{Type: "sustaining", Amount: 500000 + 211052},
		{AccretionOfMineClosureLiability: 48000},
	}

	// Calculate Actual DataSet
	apiActual := calc.CalculateDataSet(
		pbrActual,
		doreActual,
		financialActual,
		opexActual,
		capexActual,
	)

	// Create Budget DataSet (simplified for test)
	pbrBudget := &data.PBRData{
		OreMinedT:             23887,
		WasteMinedT:           293580,
		DevelopmentsM:         648,
		TotalTonnesProcessed:  35075,
		FeedGradeSilverGpt:    253.94,
		FeedGradeGoldGpt:      5.00,
		RecoveryRateSilverPct: 93.22,
		RecoveryRateGoldPct:   93.97,
	}

	dsBudget := calc.CalculateDataSet(pbrBudget, nil, nil, nil, nil)
	productionBudget := dsBudget.Production

	doreBudget := &data.DoreData{
		DoreProducedOz:       productionBudget.DoreProductionOz,
		SilverGradePct:       (productionBudget.TotalProductionSilverOz / productionBudget.DoreProductionOz) * 100,
		GoldGradePct:         (productionBudget.TotalProductionGoldOz / productionBudget.DoreProductionOz) * 100,
		PBRPriceSilver:       24.80,
		PBRPriceGold:         2050,
		RealizedPriceSilver:  24.60,
		RealizedPriceGold:    2025,
		SilverAdjustmentOz:   12,
		GoldAdjustmentOz:     6,
		AgDeductionsPct:      2.3,
		AuDeductionsPct:      1.4,
		TreatmentCharge:      5100,
		RefiningDeductionsAu: 1250,
	}

	financialBudget := &data.FinancialData{
		ShippingSelling:     120807,
		SalesTaxesRoyalties: 1230132,
	}

	opexBudget := []*data.OPEXData{
		{CostCenter: "Mine", Amount: 8240426},
		{CostCenter: "Processing", Amount: 3346617},
		{CostCenter: "G&A", Amount: 4365089},
		{Subcategory: "Inventory Variations", Amount: 603581},
	}

	capexBudget := []*data.CAPEXData{
		{Type: "sustaining", Amount: 1461635},
	}

	apiBudget := calc.CalculateDataSet(
		pbrBudget,
		doreBudget,
		financialBudget,
		opexBudget,
		capexBudget,
	)

	// Validate month/year alignment
	if reference.Month < 1 || reference.Month > 12 {
		t.Fatalf("Invalid month in reference: %d (expected 1-12)", reference.Month)
	}
	
	// Reconcile with tolerance of 1% or $1, whichever is larger
	tolerance := 0.01 // 1%
	result := Reconcile(apiActual, apiBudget, reference, tolerance)

	// Generate and print diff report
	report := GenerateDiffReport(result)
	t.Log(report)

	// Fail test if there are mismatches with clear output
	if len(result.Mismatches) > 0 {
		t.Errorf("Reconciliation FAILED: %d mismatches found (out of %d total metrics, %.1f%% match rate)",
			len(result.Mismatches), result.Summary.TotalMetrics, result.Summary.MatchRate)
		
		// List all mismatched metrics
		t.Log("\n=== MISMATCHED METRICS ===")
		for i, m := range result.Mismatches {
			t.Logf("%d. %s [%s]: Expected %.2f, Got %.2f, Diff: %.2f (%.2f%%)",
				i+1, m.MetricName, m.Category, m.ExpectedValue, m.ActualValue, m.Difference, m.DifferencePct)
		}
		
		// Show summary statistics
		t.Logf("\nMax difference: %.2f", result.Summary.MaxDifference)
		t.Logf("Max difference %%: %.2f%%", result.Summary.MaxDifferencePct)
	}
}
