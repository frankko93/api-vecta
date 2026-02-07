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
