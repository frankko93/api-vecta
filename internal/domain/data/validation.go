package data

import (
	"context"
	"fmt"
)

// CrossFileValidationError represents a validation error across multiple files
type CrossFileValidationError struct {
	Type        string   `json:"type"`         // "month_alignment", "year_mismatch", "missing_dependency"
	Message     string   `json:"message"`
	AffectedFiles []string `json:"affected_files"`
	Months      []int    `json:"months,omitempty"`
	Year        int      `json:"year,omitempty"`
}

// Error implements error interface
func (e *CrossFileValidationError) Error() string {
	return e.Message
}

// ValidateMonthAlignment validates that all data types have data for the same months
func ValidateMonthAlignment(ctx context.Context, repo Repository, companyID int64, year int, dataType string, version int) error {
	// Get months for each data type
	pbrMonths, err := getMonthsForPBR(ctx, repo, companyID, year, dataType, version)
	if err != nil {
		return fmt.Errorf("failed to get PBR months: %w", err)
	}

	doreMonths, err := getMonthsForDore(ctx, repo, companyID, year, dataType, version)
	if err != nil {
		return fmt.Errorf("failed to get Dore months: %w", err)
	}

	financialMonths, err := getMonthsForFinancial(ctx, repo, companyID, year, dataType, version)
	if err != nil {
		return fmt.Errorf("failed to get Financial months: %w", err)
	}

	opexMonths, err := getMonthsForOPEX(ctx, repo, companyID, year, dataType, version)
	if err != nil {
		return fmt.Errorf("failed to get OPEX months: %w", err)
	}

	capexMonths, err := getMonthsForCAPEX(ctx, repo, companyID, year, dataType, version)
	if err != nil {
		return fmt.Errorf("failed to get CAPEX months: %w", err)
	}

	// Find missing months in each file type
	allMonths := make(map[int]bool)
	for _, m := range pbrMonths {
		allMonths[m] = true
	}
	for _, m := range doreMonths {
		allMonths[m] = true
	}
	for _, m := range financialMonths {
		allMonths[m] = true
	}
	for _, m := range opexMonths {
		allMonths[m] = true
	}
	for _, m := range capexMonths {
		allMonths[m] = true
	}

	// Check for missing months in each file type
	var missingMonths []int
	var affectedFiles []string

	// Check PBR
	for month := range allMonths {
		if !contains(pbrMonths, month) {
			missingMonths = append(missingMonths, month)
			if !containsString(affectedFiles, "PBR") {
				affectedFiles = append(affectedFiles, "PBR")
			}
		}
	}

	// Check Dore (requires PBR, so check if PBR exists for Dore months)
	for _, month := range doreMonths {
		if !contains(pbrMonths, month) {
			missingMonths = append(missingMonths, month)
			if !containsString(affectedFiles, "Dore") {
				affectedFiles = append(affectedFiles, "Dore")
			}
			if !containsString(affectedFiles, "PBR") {
				affectedFiles = append(affectedFiles, "PBR")
			}
		}
	}

	// Check Financial
	for month := range allMonths {
		if !contains(financialMonths, month) {
			missingMonths = append(missingMonths, month)
			if !containsString(affectedFiles, "Financial") {
				affectedFiles = append(affectedFiles, "Financial")
			}
		}
	}

	// Check OPEX
	for month := range allMonths {
		if !contains(opexMonths, month) {
			missingMonths = append(missingMonths, month)
			if !containsString(affectedFiles, "OPEX") {
				affectedFiles = append(affectedFiles, "OPEX")
			}
		}
	}

	// Check CAPEX
	for month := range allMonths {
		if !contains(capexMonths, month) {
			missingMonths = append(missingMonths, month)
			if !containsString(affectedFiles, "CAPEX") {
				affectedFiles = append(affectedFiles, "CAPEX")
			}
		}
	}

	if len(missingMonths) > 0 {
		return &CrossFileValidationError{
			Type:         "month_alignment",
			Message:      fmt.Sprintf("Data files are not aligned: missing months %v in files: %v", missingMonths, affectedFiles),
			AffectedFiles: affectedFiles,
			Months:       missingMonths,
			Year:         year,
		}
	}

	return nil
}

// ValidateYearConsistency validates that all data is for the same year
func ValidateYearConsistency(ctx context.Context, repo Repository, companyID int64, dataType string, version int) error {
	years, err := getYearsForAllDataTypes(ctx, repo, companyID, dataType, version)
	if err != nil {
		return fmt.Errorf("failed to get years: %w", err)
	}

	if len(years) == 0 {
		return nil // No data, nothing to validate
	}

	// Check if all years are the same
	firstYear := years[0]
	for _, year := range years {
		if year != firstYear {
			return &CrossFileValidationError{
				Type:         "year_mismatch",
				Message:      fmt.Sprintf("Data files contain different years: found years %v", years),
				AffectedFiles: []string{"All"},
				Year:         firstYear,
			}
		}
	}

	return nil
}

// Helper functions to get months for each data type

func getMonthsForPBR(ctx context.Context, repo Repository, companyID int64, year int, dataType string, version int) ([]int, error) {
	pbrList, err := repo.ListPBRData(ctx, companyID, year, dataType, version)
	if err != nil {
		return nil, err
	}

	months := make(map[int]bool)
	for _, pbr := range pbrList {
		months[int(pbr.Date.Month())] = true
	}

	return mapKeysToSlice(months), nil
}

func getMonthsForDore(ctx context.Context, repo Repository, companyID int64, year int, dataType string, version int) ([]int, error) {
	doreList, err := repo.ListDoreData(ctx, companyID, year, dataType, version)
	if err != nil {
		return nil, err
	}

	months := make(map[int]bool)
	for _, dore := range doreList {
		months[int(dore.Date.Month())] = true
	}

	return mapKeysToSlice(months), nil
}

func getMonthsForFinancial(ctx context.Context, repo Repository, companyID int64, year int, dataType string, version int) ([]int, error) {
	financialList, err := repo.ListFinancialData(ctx, companyID, year, dataType, version)
	if err != nil {
		return nil, err
	}

	months := make(map[int]bool)
	for _, financial := range financialList {
		months[int(financial.Date.Month())] = true
	}

	return mapKeysToSlice(months), nil
}

func getMonthsForOPEX(ctx context.Context, repo Repository, companyID int64, year int, dataType string, version int) ([]int, error) {
	opexList, err := repo.ListOPEXData(ctx, companyID, year, dataType, version)
	if err != nil {
		return nil, err
	}

	months := make(map[int]bool)
	for _, opex := range opexList {
		months[int(opex.Date.Month())] = true
	}

	return mapKeysToSlice(months), nil
}

func getMonthsForCAPEX(ctx context.Context, repo Repository, companyID int64, year int, dataType string, version int) ([]int, error) {
	capexList, err := repo.ListCAPEXData(ctx, companyID, year, dataType, version)
	if err != nil {
		return nil, err
	}

	months := make(map[int]bool)
	for _, capex := range capexList {
		months[int(capex.Date.Month())] = true
	}

	return mapKeysToSlice(months), nil
}

func getYearsForAllDataTypes(ctx context.Context, repo Repository, companyID int64, dataType string, version int) ([]int, error) {
	years := make(map[int]bool)

	// Get years from each data type
	pbrList, _ := repo.ListPBRData(ctx, companyID, 0, dataType, version) // year=0 means all years
	for _, pbr := range pbrList {
		years[pbr.Date.Year()] = true
	}

	doreList, _ := repo.ListDoreData(ctx, companyID, 0, dataType, version)
	for _, dore := range doreList {
		years[dore.Date.Year()] = true
	}

	financialList, _ := repo.ListFinancialData(ctx, companyID, 0, dataType, version)
	for _, financial := range financialList {
		years[financial.Date.Year()] = true
	}

	opexList, _ := repo.ListOPEXData(ctx, companyID, 0, dataType, version)
	for _, opex := range opexList {
		years[opex.Date.Year()] = true
	}

	capexList, _ := repo.ListCAPEXData(ctx, companyID, 0, dataType, version)
	for _, capex := range capexList {
		years[capex.Date.Year()] = true
	}

	return mapKeysToSlice(years), nil
}

// Helper utility functions

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func containsString(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func mapKeysToSlice(m map[int]bool) []int {
	slice := make([]int, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}

// ValidateDoreDependencies validates that Dore data has corresponding PBR data
func ValidateDoreDependencies(ctx context.Context, repo Repository, companyID int64, year int, dataType string, version int) error {
	doreList, err := repo.ListDoreData(ctx, companyID, year, dataType, version)
	if err != nil {
		return fmt.Errorf("failed to get Dore data: %w", err)
	}

	if len(doreList) == 0 {
		return nil // No Dore data, nothing to validate
	}

	pbrList, err := repo.ListPBRData(ctx, companyID, year, dataType, version)
	if err != nil {
		return fmt.Errorf("failed to get PBR data: %w", err)
	}

	// Create map of PBR dates
	pbrDates := make(map[string]bool)
	for _, pbr := range pbrList {
		dateKey := pbr.Date.Format("2006-01-02")
		pbrDates[dateKey] = true
	}

	// Check each Dore date has corresponding PBR
	var missingDates []string
	for _, dore := range doreList {
		dateKey := dore.Date.Format("2006-01-02")
		if !pbrDates[dateKey] {
			missingDates = append(missingDates, dateKey)
		}
	}

	if len(missingDates) > 0 {
		return &CrossFileValidationError{
			Type:         "missing_dependency",
			Message:      fmt.Sprintf("Dore data requires PBR data for the same dates. Missing PBR for dates: %v", missingDates),
			AffectedFiles: []string{"Dore", "PBR"},
			Year:         year,
		}
	}

	return nil
}
