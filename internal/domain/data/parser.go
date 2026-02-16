package data

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gmhafiz/go8/internal/domain/config"
)

// Helper functions for parsing CSV

// Helper functions for parsing CSV

func parseDate(value string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %s", value)
	}
	return t, nil
}

// parseFloat robustly parses numeric values handling:
// - Commas as thousand separators: "24,859" -> 24859
// - Currency symbols: "$ 123.45" or "$123.45" -> 123.45
// - Percentage signs: "94.01%" -> 94.01
// - Negative values in parentheses: "(30,989)" -> -30989
// - Whitespace: " 123 " -> 123
// If required=true, "-" or empty values return an error. If required=false, they return 0.
func parseFloat(value string, required bool) (float64, error) {
	value = strings.TrimSpace(value)

	// Check for empty/dash values (before any processing)
	isEmpty := value == "" || value == "-" || value == "$ -" || value == "$ -   "
	if isEmpty {
		if required {
			return 0, fmt.Errorf("required field cannot be empty or dash")
		}
		return 0, nil // Optional field: empty/dash represents zero
	}

	// Handle negative values in parentheses FIRST: (123) -> -123
	// This must be done before removing currency symbols
	// Handle cases like " $ (202) " or "$ (8,537,997)"
	isNegative := false
	trimmedValue := strings.TrimSpace(value)

	// Find first ( and last ) accounting for spaces
	openIdx := strings.Index(trimmedValue, "(")
	closeIdx := strings.LastIndex(trimmedValue, ")")

	if openIdx >= 0 && closeIdx > openIdx {
		// Extract content between parentheses
		content := trimmedValue[openIdx+1 : closeIdx]
		content = strings.TrimSpace(content)

		// Check if there's meaningful content (not just spaces)
		if content != "" {
			isNegative = true
			value = content
		}
	}

	// Remove currency symbols ($, USD, etc.) - AFTER handling parentheses
	value = strings.TrimPrefix(value, "$")
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "USD")
	value = strings.TrimSpace(value)

	// Remove percentage sign
	value = strings.TrimSuffix(value, "%")
	value = strings.TrimSpace(value)

	// Remove commas (thousand separators)
	value = strings.ReplaceAll(value, ",", "")

	// Parse as float
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", value)
	}

	// Apply negative sign if it was in parentheses
	if isNegative {
		f = -f
	}

	return f, nil
}

func readCSV(fileContent []byte, expectedHeaders []string) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(fileContent))

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, ErrInvalidCSVFormat
	}

	headers := records[0]

	if len(headers) != len(expectedHeaders) {
		return nil, fmt.Errorf("expected %d columns, got %d", len(expectedHeaders), len(headers))
	}

	for i, expected := range expectedHeaders {
		if strings.TrimSpace(headers[i]) != expected {
			return nil, fmt.Errorf("header mismatch at column %d: expected '%s', got '%s'", i+1, expected, headers[i])
		}
	}

	return records[1:], nil
}

func validateRow(row []string, expectedColumns int, rowNum int) error {
	if len(row) != expectedColumns {
		return fmt.Errorf("row %d: expected %d columns, got %d", rowNum, expectedColumns, len(row))
	}
	return nil
}

// readCSVForDates reads CSV and returns rows (without validation) - used to get dates before full parsing
func readCSVForDates(fileContent []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(fileContent))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, ErrInvalidCSVFormat
	}
	// Return data rows (skip header)
	return records[1:], nil
}

// Parsers for each data type

var productionHeaders = []string{"date", "mineral_code", "quantity", "unit"}

func parseProductionCSV(fileContent []byte, companyID, userID int64, dataType string, version int, description string, mineralMap map[string]int) ([]*ProductionData, []ValidationError) {
	rows, err := readCSV(fileContent, productionHeaders)
	if err != nil {
		return nil, []ValidationError{{Row: 0, Error: err.Error()}}
	}

	var records []*ProductionData
	var errors []ValidationError

	for i, row := range rows {
		rowNum := i + 2

		if err := validateRow(row, len(productionHeaders), rowNum); err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Error: err.Error()})
			continue
		}

		date, err := parseDate(row[0])
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "date", Error: err.Error()})
			continue
		}

		mineralCode := strings.TrimSpace(row[1])
		mineralID, exists := mineralMap[mineralCode]
		if !exists {
			errors = append(errors, ValidationError{Row: rowNum, Column: "mineral_code", Error: fmt.Sprintf("mineral not found: %s", mineralCode)})
			continue
		}

		quantity, err := parseFloat(row[2], true) // Required
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "quantity", Error: err.Error()})
			continue
		}
		if quantity <= 0 {
			errors = append(errors, ValidationError{Row: rowNum, Column: "quantity", Error: "must be greater than 0"})
			continue
		}

		unit := config.UnitOfMeasure(strings.TrimSpace(row[3]))
		if !unit.IsValid() {
			errors = append(errors, ValidationError{Row: rowNum, Column: "unit", Error: fmt.Sprintf("invalid unit: %s", row[3])})
			continue
		}

		records = append(records, &ProductionData{
			CompanyID:   companyID,
			Date:        date,
			MineralID:   mineralID,
			Quantity:    quantity,
			Unit:        string(unit),
			DataType:    dataType,
			Version:     version,
			Description: description,
			CreatedBy:   userID,
		})
	}

	return records, errors
}

var doreHeaders = []string{
	"date",
	"pbr_price_silver", "pbr_price_gold", "realized_price_silver", "realized_price_gold",
	"silver_adjustment_oz", "gold_adjustment_oz", "ag_deductions_pct", "au_deductions_pct",
	"treatment_charge", "refining_deductions_au", "streaming",
}

// parseDoreCSV parses Dore CSV and calculates production from PBR data
// PBR data is required to calculate dore_produced_oz, silver_grade_pct, and gold_grade_pct
func parseDoreCSV(fileContent []byte, companyID, userID int64, dataType string, version int, description string, pbrMap map[string]*PBRData) ([]*DoreData, []ValidationError) {
	rows, err := readCSV(fileContent, doreHeaders)
	if err != nil {
		return nil, []ValidationError{{Row: 0, Error: err.Error()}}
	}

	var records []*DoreData
	var errors []ValidationError

	for i, row := range rows {
		rowNum := i + 2

		if err := validateRow(row, len(doreHeaders), rowNum); err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Error: err.Error()})
			continue
		}

		date, err := parseDate(row[0])
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "date", Error: err.Error()})
			continue
		}

		// Get PBR data for this date
		dateKey := date.Format("2006-01-02")
		pbr, exists := pbrMap[dateKey]
		if !exists {
			errors = append(errors, ValidationError{
				Row:    rowNum,
				Column: "date",
				Error:  fmt.Sprintf("PBR data not found for date %s. Please import PBR data first.", dateKey),
			})
			continue
		}

		// Calculate production from PBR
		// Formula: Feed Grade (g/t) * Tonnes Processed * Recovery Rate / 31.1035 (grams per oz)
		silverOz := pbr.FeedGradeSilverGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateSilverPct / 100) / 31.1035
		goldOz := pbr.FeedGradeGoldGpt * pbr.TotalTonnesProcessed * (pbr.RecoveryRateGoldPct / 100) / 31.1035
		doreProducedOz := silverOz + goldOz

		// Calculate grades
		var silverGradePct, goldGradePct float64
		if doreProducedOz > 0 {
			silverGradePct = (silverOz / doreProducedOz) * 100
			goldGradePct = (goldOz / doreProducedOz) * 100
		}

		// Parse remaining values from CSV (refining-specific data)
		// Most Dore fields are required, streaming is optional (defaults to 0)
		values := make([]float64, 11) // 11 fields after date
		for j := 1; j < len(doreHeaders); j++ {
			// Streaming (last field) is optional and can be negative
			isOptional := doreHeaders[j] == "streaming"
			values[j-1], err = parseFloat(row[j], !isOptional)
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: doreHeaders[j], Error: err.Error()})
				break
			}
		}
		if err != nil {
			continue
		}

		records = append(records, &DoreData{
			CompanyID:            companyID,
			Date:                 date,
			DoreProducedOz:       doreProducedOz, // Calculated from PBR
			SilverGradePct:       silverGradePct, // Calculated from PBR
			GoldGradePct:         goldGradePct,   // Calculated from PBR
			PBRPriceSilver:       values[0],      // pbr_price_silver
			PBRPriceGold:         values[1],      // pbr_price_gold
			RealizedPriceSilver:  values[2],      // realized_price_silver
			RealizedPriceGold:    values[3],      // realized_price_gold
			SilverAdjustmentOz:   values[4],      // silver_adjustment_oz
			GoldAdjustmentOz:     values[5],      // gold_adjustment_oz
			AgDeductionsPct:      values[6],      // ag_deductions_pct
			AuDeductionsPct:      values[7],      // au_deductions_pct
			TreatmentCharge:      values[8],      // treatment_charge
			RefiningDeductionsAu: values[9],      // refining_deductions_au
			Streaming:            values[10],     // streaming (can be negative)
			DataType:             dataType,
			Version:              version,
			Description:          description,
			CreatedBy:            userID,
		})
	}

	return records, errors
}

var pbrHeaders = []string{
	"date", "ore_mined_t", "waste_mined_t", "developments_m",
	"total_tonnes_processed", "feed_grade_silver_gpt", "feed_grade_gold_gpt",
	"recovery_rate_silver_pct", "recovery_rate_gold_pct",
}

func parsePBRCSV(fileContent []byte, companyID, userID int64, dataType string, version int, description string) ([]*PBRData, []ValidationError) {
	rows, err := readCSV(fileContent, pbrHeaders)
	if err != nil {
		return nil, []ValidationError{{Row: 0, Error: err.Error()}}
	}

	var records []*PBRData
	var errors []ValidationError

	for i, row := range rows {
		rowNum := i + 2

		if err := validateRow(row, len(pbrHeaders), rowNum); err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Error: err.Error()})
			continue
		}

		date, err := parseDate(row[0])
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "date", Error: err.Error()})
			continue
		}

		// All PBR fields are required
		values := make([]float64, 8)
		for j := 1; j < 9; j++ {
			values[j-1], err = parseFloat(row[j], true) // Required
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: pbrHeaders[j], Error: err.Error()})
				break
			}
		}
		if err != nil {
			continue
		}

		records = append(records, &PBRData{
			CompanyID:             companyID,
			Date:                  date,
			OreMinedT:             values[0],
			WasteMinedT:           values[1],
			DevelopmentsM:         values[2],
			TotalTonnesProcessed:  values[3],
			FeedGradeSilverGpt:    values[4],
			FeedGradeGoldGpt:      values[5],
			RecoveryRateSilverPct: values[6],
			RecoveryRateGoldPct:   values[7],
			DataType:              dataType,
			Version:               version,
			Description:           description,
			CreatedBy:             userID,
		})
	}

	return records, errors
}

var opexHeaders = []string{"date", "cost_center", "subcategory", "expense_type", "amount", "currency"}

func parseOPEXCSV(fileContent []byte, companyID, userID int64, dataType string, version int, description string) ([]*OPEXData, []ValidationError) {
	rows, err := readCSV(fileContent, opexHeaders)
	if err != nil {
		return nil, []ValidationError{{Row: 0, Error: err.Error()}}
	}

	var records []*OPEXData
	var errors []ValidationError

	for i, row := range rows {
		rowNum := i + 2

		if err := validateRow(row, len(opexHeaders), rowNum); err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Error: err.Error()})
			continue
		}

		date, err := parseDate(row[0])
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "date", Error: err.Error()})
			continue
		}

		costCenter := CostCenter(strings.TrimSpace(row[1]))
		if !costCenter.IsValid() {
			errors = append(errors, ValidationError{Row: rowNum, Column: "cost_center", Error: fmt.Sprintf("invalid cost center: %s", row[1])})
			continue
		}

		subcategory := strings.TrimSpace(row[2])
		if subcategory == "" {
			errors = append(errors, ValidationError{Row: rowNum, Column: "subcategory", Error: "subcategory is required"})
			continue
		}

		expenseType := ExpenseType(strings.TrimSpace(row[3]))
		if !expenseType.IsValid() {
			errors = append(errors, ValidationError{Row: rowNum, Column: "expense_type", Error: fmt.Sprintf("invalid expense type: %s", row[3])})
			continue
		}

		amount, err := parseFloat(row[4], true) // Required
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "amount", Error: err.Error()})
			continue
		}
		if amount < 0 {
			errors = append(errors, ValidationError{Row: rowNum, Column: "amount", Error: "amount cannot be negative"})
			continue
		}

		currency := Currency(strings.TrimSpace(row[5]))
		if !currency.IsValid() {
			errors = append(errors, ValidationError{Row: rowNum, Column: "currency", Error: fmt.Sprintf("invalid currency: %s", row[5])})
			continue
		}

		records = append(records, &OPEXData{
			CompanyID:   companyID,
			Date:        date,
			CostCenter:  string(costCenter),
			Subcategory: subcategory,
			ExpenseType: string(expenseType),
			Amount:      amount,
			Currency:    string(currency),
			DataType:    dataType,
			Version:     version,
			Description: description,
			CreatedBy:   userID,
		})
	}

	return records, errors
}

var capexHeaders = []string{"date", "category", "car_number", "project_name", "type", "amount", "accretion_of_mine_closure_liability", "currency"}

func parseCAPEXCSV(fileContent []byte, companyID, userID int64, dataType string, version int, description string) ([]*CAPEXData, []ValidationError) {
	rows, err := readCSV(fileContent, capexHeaders)
	if err != nil {
		return nil, []ValidationError{{Row: 0, Error: err.Error()}}
	}

	var records []*CAPEXData
	var errors []ValidationError

	for i, row := range rows {
		rowNum := i + 2

		if err := validateRow(row, len(capexHeaders), rowNum); err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Error: err.Error()})
			continue
		}

		date, err := parseDate(row[0])
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "date", Error: err.Error()})
			continue
		}

		category := strings.TrimSpace(row[1])
		if category == "" {
			errors = append(errors, ValidationError{Row: rowNum, Column: "category", Error: "category is required"})
			continue
		}

		carNumber := strings.TrimSpace(row[2])

		projectName := strings.TrimSpace(row[3])
		if projectName == "" {
			// Use category as fallback for summary/total rows without a specific project name
			projectName = category
		}

		capexType := CapexType(strings.TrimSpace(row[4]))
		if !capexType.IsValid() {
			errors = append(errors, ValidationError{Row: rowNum, Column: "type", Error: fmt.Sprintf("invalid type: %s", row[4])})
			continue
		}

		amount, err := parseFloat(row[5], true) // Required
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "amount", Error: err.Error()})
			continue
		}
		// CAPEX allows negative amounts for accounting adjustments/reversals

		// Parse accretion_of_mine_closure_liability (optional, defaults to 0)
		accretion := 0.0
		if strings.TrimSpace(row[6]) != "" {
			accretion, err = parseFloat(row[6], false) // Optional
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: "accretion_of_mine_closure_liability", Error: err.Error()})
				continue
			}
			if accretion < 0 {
				errors = append(errors, ValidationError{Row: rowNum, Column: "accretion_of_mine_closure_liability", Error: "accretion cannot be negative"})
				continue
			}
		}

		currency := Currency(strings.TrimSpace(row[7]))
		if !currency.IsValid() {
			errors = append(errors, ValidationError{Row: rowNum, Column: "currency", Error: fmt.Sprintf("invalid currency: %s", row[7])})
			continue
		}

		records = append(records, &CAPEXData{
			CompanyID:                       companyID,
			Date:                            date,
			Category:                        category,
			CARNumber:                       carNumber,
			ProjectName:                     projectName,
			Type:                            string(capexType),
			Amount:                          amount,
			AccretionOfMineClosureLiability: accretion,
			Currency:                        string(currency),
			DataType:                        dataType,
			Version:                         version,
			Description:                     description,
			CreatedBy:                       userID,
		})
	}

	return records, errors
}

var revenueHeaders = []string{"date", "mineral_code", "quantity_sold", "unit_price", "currency"}

func parseRevenueCSV(fileContent []byte, companyID, userID int64, dataType string, version int, description string, mineralMap map[string]int) ([]*RevenueData, []ValidationError) {
	rows, err := readCSV(fileContent, revenueHeaders)
	if err != nil {
		return nil, []ValidationError{{Row: 0, Error: err.Error()}}
	}

	var records []*RevenueData
	var errors []ValidationError

	for i, row := range rows {
		rowNum := i + 2

		if err := validateRow(row, len(revenueHeaders), rowNum); err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Error: err.Error()})
			continue
		}

		date, err := parseDate(row[0])
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "date", Error: err.Error()})
			continue
		}

		mineralCode := strings.TrimSpace(row[1])
		mineralID, exists := mineralMap[mineralCode]
		if !exists {
			errors = append(errors, ValidationError{Row: rowNum, Column: "mineral_code", Error: fmt.Sprintf("mineral not found: %s", mineralCode)})
			continue
		}

		quantitySold, err := parseFloat(row[2], true) // Required
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "quantity_sold", Error: err.Error()})
			continue
		}
		if quantitySold <= 0 {
			errors = append(errors, ValidationError{Row: rowNum, Column: "quantity_sold", Error: "must be greater than 0"})
			continue
		}

		unitPrice, err := parseFloat(row[3], true) // Required
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "unit_price", Error: err.Error()})
			continue
		}
		if unitPrice <= 0 {
			errors = append(errors, ValidationError{Row: rowNum, Column: "unit_price", Error: "must be greater than 0"})
			continue
		}

		currency := Currency(strings.TrimSpace(row[4]))
		if !currency.IsValid() {
			errors = append(errors, ValidationError{Row: rowNum, Column: "currency", Error: fmt.Sprintf("invalid currency: %s", row[4])})
			continue
		}

		records = append(records, &RevenueData{
			CompanyID:    companyID,
			Date:         date,
			MineralID:    mineralID,
			QuantitySold: quantitySold,
			UnitPrice:    unitPrice,
			Currency:     string(currency),
			DataType:     dataType,
			Version:      version,
			Description:  description,
			CreatedBy:    userID,
		})
	}

	return records, errors
}

// New format: split sales_taxes and royalties, plus other_sales_deductions
var financialHeaders = []string{"date", "shipping_selling", "sales_taxes", "royalties", "other_sales_deductions", "other_adjustments"}

// Legacy format: combined sales_taxes_royalties (backward compatibility)
var financialHeadersLegacy = []string{"date", "shipping_selling", "sales_taxes_royalties", "other_adjustments"}

func parseFinancialCSV(fileContent []byte, companyID, userID int64, dataType string, version int, description string) ([]*FinancialData, []ValidationError) {
	// Try new format first, fall back to legacy format
	rows, err := readCSV(fileContent, financialHeaders)
	useLegacy := false
	if err != nil {
		// Try legacy format with combined sales_taxes_royalties
		rows, err = readCSV(fileContent, financialHeadersLegacy)
		if err != nil {
			return nil, []ValidationError{{Row: 0, Error: err.Error()}}
		}
		useLegacy = true
	}

	var records []*FinancialData
	var errors []ValidationError

	for i, row := range rows {
		rowNum := i + 2

		if useLegacy {
			if err := validateRow(row, len(financialHeadersLegacy), rowNum); err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Error: err.Error()})
				continue
			}
		} else {
			if err := validateRow(row, len(financialHeaders), rowNum); err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Error: err.Error()})
				continue
			}
		}

		date, err := parseDate(row[0])
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "date", Error: err.Error()})
			continue
		}

		shippingSelling, err := parseFloat(row[1], true) // Required
		if err != nil {
			errors = append(errors, ValidationError{Row: rowNum, Column: "shipping_selling", Error: err.Error()})
			continue
		}

		var salesTaxes, royalties, otherSalesDeductions, otherAdjustments float64

		if useLegacy {
			// Legacy format: sales_taxes_royalties (combined) -> map to sales_taxes, royalties=0
			salesTaxesRoyalties, err := parseFloat(row[2], true) // Required
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: "sales_taxes_royalties", Error: err.Error()})
				continue
			}
			salesTaxes = salesTaxesRoyalties
			royalties = 0
			otherSalesDeductions = 0

			otherAdjustments, err = parseFloat(row[3], false) // Optional
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: "other_adjustments", Error: err.Error()})
				continue
			}
		} else {
			// New format: separate fields
			salesTaxes, err = parseFloat(row[2], true) // Required
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: "sales_taxes", Error: err.Error()})
				continue
			}

			royalties, err = parseFloat(row[3], true) // Required
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: "royalties", Error: err.Error()})
				continue
			}

			otherSalesDeductions, err = parseFloat(row[4], false) // Optional
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: "other_sales_deductions", Error: err.Error()})
				continue
			}

			otherAdjustments, err = parseFloat(row[5], false) // Optional
			if err != nil {
				errors = append(errors, ValidationError{Row: rowNum, Column: "other_adjustments", Error: err.Error()})
				continue
			}
		}

		records = append(records, &FinancialData{
			CompanyID:            companyID,
			Date:                 date,
			ShippingSelling:      shippingSelling,
			SalesTaxes:           salesTaxes,
			Royalties:            royalties,
			OtherSalesDeductions: otherSalesDeductions,
			OtherAdjustments:     otherAdjustments,
			Currency:             "USD",
			DataType:             dataType,
			Version:              version,
			Description:          description,
			CreatedBy:            userID,
		})
	}

	return records, errors
}
