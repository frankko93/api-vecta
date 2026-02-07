package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseProductionCSV_Success(t *testing.T) {
	csvContent := buildProductionCSV([]string{
		validProductionRow,
		"2024-01-16,AG,2300,kilograms",
	})

	records, errors := parseProductionCSV(csvContent, testCompanyID, testUserID, "actual", testVersion, testDescription, getTestMineralMap())

	assert.Empty(t, errors)
	assert.Len(t, records, 2)
	assert.Equal(t, 1, records[0].MineralID)
	assert.Equal(t, 150.5, records[0].Quantity)
	assert.Equal(t, "kilograms", records[0].Unit)
	assert.Equal(t, "actual", records[0].DataType)
}

func TestParseProductionCSV_InvalidMineralCode(t *testing.T) {
	csvContent := buildProductionCSV([]string{
		"2024-01-15,XYZ,150.5,kilograms",
	})

	_, errors := parseProductionCSV(csvContent, testCompanyID, testUserID, "actual", testVersion, testDescription, getTestMineralMap())

	assert.Len(t, errors, 1)
	assert.Equal(t, 2, errors[0].Row)
	assert.Contains(t, errors[0].Error, "mineral not found")
}

func TestParseProductionCSV_InvalidQuantity(t *testing.T) {
	csvContent := buildProductionCSV([]string{
		"2024-01-15,AU,-10,kilograms",
	})

	_, errors := parseProductionCSV(csvContent, testCompanyID, testUserID, "actual", testVersion, testDescription, getTestMineralMap())

	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error, "must be greater than 0")
}

func TestParseProductionCSV_InvalidHeaders(t *testing.T) {
	csvContent := []byte("date,wrong_header,quantity,unit\n2024-01-15,AU,150.5,kilograms")

	_, errors := parseProductionCSV(csvContent, testCompanyID, testUserID, "actual", testVersion, testDescription, getTestMineralMap())

	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error, "header mismatch")
}

func TestParsePBRCSV_Success(t *testing.T) {
	csvContent := buildPBRCSV([]string{
		validPBRRow,
	})

	records, errors := parsePBRCSV(csvContent, testCompanyID, testUserID, "budget", testVersion, testDescription)

	assert.Empty(t, errors)
	assert.Len(t, records, 1)
	assert.Equal(t, 24859.0, records[0].OreMinedT)
	assert.Equal(t, "budget", records[0].DataType)
}

func TestParseOPEXCSV_Success(t *testing.T) {
	csvContent := buildOPEXCSV([]string{
		validOPEXRow,
		"2024-01-15,Processing,CO General Operating,Materials,20000,USD",
	})

	records, errors := parseOPEXCSV(csvContent, testCompanyID, testUserID, "actual", testVersion, testDescription)

	assert.Empty(t, errors)
	assert.Len(t, records, 2)
	assert.Equal(t, "Mine", records[0].CostCenter)
	assert.Equal(t, 50000.0, records[0].Amount)
}

func TestParseOPEXCSV_InvalidCostCenter(t *testing.T) {
	csvContent := buildOPEXCSV([]string{
		"2024-01-15,InvalidCenter,Drilling,Labour,50000,USD",
	})

	_, errors := parseOPEXCSV(csvContent, testCompanyID, testUserID, "actual", testVersion, testDescription)

	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error, "invalid cost center")
}

func TestParseFinancialCSV_Success(t *testing.T) {
	csvContent := buildFinancialCSV([]string{
		validFinancialRow,
	})

	records, validationErrors := parseFinancialCSV(csvContent, testCompanyID, testUserID, "actual", testVersion, testDescription)

	assert.Empty(t, validationErrors)
	assert.Len(t, records, 1)
	assert.Equal(t, -202.0, records[0].ShippingSelling)
	assert.Equal(t, 465867.0, records[0].SalesTaxesRoyalties)
	assert.Equal(t, "USD", records[0].Currency)
}

func TestParseDateFormat(t *testing.T) {
	// Valid date
	date, err := parseDate("2024-01-15")
	assert.NoError(t, err)
	assert.Equal(t, 2024, date.Year())
	assert.Equal(t, time.Month(1), date.Month())
	assert.Equal(t, 15, date.Day())

	// Invalid date format
	_, err = parseDate("15/01/2024")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date format")

	// Invalid date
	_, err = parseDate("2024-13-45")
	assert.Error(t, err)
}

// Test robust numeric parsing
func TestParseFloat_RobustParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		// Basic numbers
		{"simple integer", "123", 123.0, false},
		{"simple float", "123.45", 123.45, false},
		{"negative", "-123", -123.0, false},

		// Numbers with commas (thousand separators)
		{"comma thousands", "24,859", 24859.0, false},
		{"comma millions", "1,234,567", 1234567.0, false},
		{"comma with decimals", "1,234.56", 1234.56, false},

		// Currency symbols
		{"dollar prefix", "$123.45", 123.45, false},
		{"dollar with space", "$ 123.45", 123.45, false},
		{"dollar with comma", "$ 24,859", 24859.0, false},
		{"dollar negative", "$ -123", -123.0, false},

		// Percentages
		{"percentage", "94.01%", 94.01, false},
		{"percentage with space", "94.01 %", 94.01, false},
		{"percentage with comma", "1,234.56%", 1234.56, false},

		// Negative in parentheses
		{"parentheses negative", "(123)", -123.0, false},
		{"parentheses with comma", "(30,989)", -30989.0, false},
		{"parentheses with dollar", "$ (123)", -123.0, false},
		{"parentheses with dollar and comma", "$ (30,989)", -30989.0, false},

		// Empty/zero values (only valid for optional fields)
		{"empty string optional", "", 0.0, false},
		{"dash optional", "-", 0.0, false},
		{"dollar dash optional", "$ -", 0.0, false},
		{"dollar dash spaces optional", "$ -   ", 0.0, false},

		// Whitespace
		{"whitespace", "  123  ", 123.0, false},
		{"whitespace with comma", "  24,859  ", 24859.0, false},

		// Complex combinations
		{"dollar comma percentage", "$ 1,234.56%", 1234.56, false},
		{"parentheses dollar comma", "$ (1,234.56)", -1234.56, false},

		// Invalid
		{"invalid text", "abc", 0.0, true},
		{"invalid mixed", "12abc", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test as optional field (allows empty/dash)
			result, err := parseFloat(tt.input, false)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.01, "input: %s", tt.input)
			}
		})
	}
}

// Test parseFloat with required=true (should reject empty/dash)
func TestParseFloat_Required(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string required", "", true},
		{"dash required", "-", true},
		{"dollar dash required", "$ -", true},
		{"dollar dash spaces required", "$ -   ", true},
		{"valid number required", "123", false},
		{"valid comma number required", "24,859", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFloat(tt.input, true)
			if tt.wantErr {
				assert.Error(t, err, "should reject empty/dash for required field: %s", tt.input)
				assert.Contains(t, err.Error(), "required field cannot be empty")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test parsing with real Summary.csv format examples
func TestParseFloat_SummaryCSVExamples(t *testing.T) {
	examples := map[string]float64{
		"24,859":         24859.0,
		"(30,989)":       -30989.0,
		"209.79":         209.79,
		"94.01%":         94.01,
		"$ 763.6":        763.6,
		"$ (8,537,997)":  -8537997.0,
		"$ -":            0.0,
		"$ -   ":         0.0,
		" $ 27,919,108 ": 27919108.0,
		" $ (202)":       -202.0,
	}

	for input, expected := range examples {
		t.Run(input, func(t *testing.T) {
			// Test as optional (allows dash)
			result, err := parseFloat(input, false)
			assert.NoError(t, err)
			assert.InDelta(t, expected, result, 0.01, "input: %s", input)
		})
	}
}
