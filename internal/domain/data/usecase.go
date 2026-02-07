package data

import (
	"context"
)

type UseCase interface {
	ImportData(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error)
	ListData(ctx context.Context, dataType DataImportType, companyID int64, year int, typeFilter string, version int) (interface{}, error)
	DeleteData(ctx context.Context, dataType DataImportType, id int64) error
}

type useCase struct {
	repo Repository
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) ImportData(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error) {
	// Validate data type
	if !req.Type.IsValid() {
		return nil, ErrInvalidDataType
	}

	// Validate company exists
	exists, err := uc.repo.CompanyExists(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCompanyNotFound
	}

	// Set default version if not provided
	if req.Version == 0 {
		req.Version = 1
	}

	var response *ImportResponse

	switch req.Type {
	case ImportProduction:
		response, err = uc.importProduction(ctx, req, userID)
	case ImportDore:
		response, err = uc.importDore(ctx, req, userID)
	case ImportPBR:
		response, err = uc.importPBR(ctx, req, userID)
	case ImportOPEX:
		response, err = uc.importOPEX(ctx, req, userID)
	case ImportCAPEX:
		response, err = uc.importCAPEX(ctx, req, userID)
	case ImportRevenue:
		response, err = uc.importRevenue(ctx, req, userID)
	case ImportFinancial:
		response, err = uc.importFinancial(ctx, req, userID)
	default:
		return nil, ErrInvalidDataType
	}

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (uc *useCase) importProduction(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error) {
	// Get mineral code map
	mineralMap, err := uc.repo.GetMineralCodeMap(ctx)
	if err != nil {
		return nil, err
	}

	// Parse CSV
	records, validationErrors := parseProductionCSV(req.File, req.CompanyID, userID, req.DataType, req.Version, req.Description, mineralMap)

	// If any validation errors, fail the entire import
	if len(validationErrors) > 0 {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    len(records) + len(validationErrors),
			RowsInserted: 0,
			RowsFailed:   len(validationErrors),
			Errors:       validationErrors,
		}, nil
	}

	// Insert all records in transaction
	err = uc.repo.InsertProductionBulk(ctx, records)
	if err != nil {
		return nil, err
	}

	return &ImportResponse{
		Success:      true,
		Type:         req.Type,
		RowsTotal:    len(records),
		RowsInserted: len(records),
		RowsFailed:   0,
		Errors:       []ValidationError{},
	}, nil
}

func (uc *useCase) importDore(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error) {
	// Get PBR data for the same year, data type, and version to calculate dore production
	// We need to parse the CSV first to get the dates, but we'll do a two-pass approach
	// First, read CSV to get dates
	rows, err := readCSVForDates(req.File)
	if err != nil {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    0,
			RowsInserted: 0,
			RowsFailed:   0,
			Errors:       []ValidationError{{Row: 0, Error: err.Error()}},
		}, nil
	}

	// Get year from first date
	if len(rows) == 0 {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    0,
			RowsInserted: 0,
			RowsFailed:   0,
			Errors:       []ValidationError{{Row: 0, Error: "CSV file is empty"}},
		}, nil
	}

	firstDate, err := parseDate(rows[0][0])
	if err != nil {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    0,
			RowsInserted: 0,
			RowsFailed:   0,
			Errors:       []ValidationError{{Row: 2, Column: "date", Error: err.Error()}},
		}, nil
	}

	year := firstDate.Year()

	// Get all PBR data for this year
	pbrList, err := uc.repo.ListPBRData(ctx, req.CompanyID, year, req.DataType, req.Version)
	if err != nil {
		return nil, err
	}

	// Create a map of PBR data by date for quick lookup
	pbrMap := make(map[string]*PBRData)
	for i := range pbrList {
		dateKey := pbrList[i].Date.Format("2006-01-02")
		pbrMap[dateKey] = pbrList[i]
	}

	// Now parse Dore CSV with PBR data
	records, validationErrors := parseDoreCSV(req.File, req.CompanyID, userID, req.DataType, req.Version, req.Description, pbrMap)

	if len(validationErrors) > 0 {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    len(records) + len(validationErrors),
			RowsInserted: 0,
			RowsFailed:   len(validationErrors),
			Errors:       validationErrors,
		}, nil
	}

	err = uc.repo.InsertDoreBulk(ctx, records)
	if err != nil {
		return nil, err
	}

	return &ImportResponse{
		Success:      true,
		Type:         req.Type,
		RowsTotal:    len(records),
		RowsInserted: len(records),
		RowsFailed:   0,
		Errors:       []ValidationError{},
	}, nil
}

func (uc *useCase) importPBR(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error) {
	records, validationErrors := parsePBRCSV(req.File, req.CompanyID, userID, req.DataType, req.Version, req.Description)

	if len(validationErrors) > 0 {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    len(records) + len(validationErrors),
			RowsInserted: 0,
			RowsFailed:   len(validationErrors),
			Errors:       validationErrors,
		}, nil
	}

	err := uc.repo.InsertPBRBulk(ctx, records)
	if err != nil {
		return nil, err
	}

	return &ImportResponse{
		Success:      true,
		Type:         req.Type,
		RowsTotal:    len(records),
		RowsInserted: len(records),
		RowsFailed:   0,
		Errors:       []ValidationError{},
	}, nil
}

func (uc *useCase) importOPEX(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error) {
	records, validationErrors := parseOPEXCSV(req.File, req.CompanyID, userID, req.DataType, req.Version, req.Description)

	if len(validationErrors) > 0 {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    len(records) + len(validationErrors),
			RowsInserted: 0,
			RowsFailed:   len(validationErrors),
			Errors:       validationErrors,
		}, nil
	}

	err := uc.repo.InsertOPEXBulk(ctx, records)
	if err != nil {
		return nil, err
	}

	return &ImportResponse{
		Success:      true,
		Type:         req.Type,
		RowsTotal:    len(records),
		RowsInserted: len(records),
		RowsFailed:   0,
		Errors:       []ValidationError{},
	}, nil
}

func (uc *useCase) importCAPEX(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error) {
	records, validationErrors := parseCAPEXCSV(req.File, req.CompanyID, userID, req.DataType, req.Version, req.Description)

	if len(validationErrors) > 0 {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    len(records) + len(validationErrors),
			RowsInserted: 0,
			RowsFailed:   len(validationErrors),
			Errors:       validationErrors,
		}, nil
	}

	err := uc.repo.InsertCAPEXBulk(ctx, records)
	if err != nil {
		return nil, err
	}

	return &ImportResponse{
		Success:      true,
		Type:         req.Type,
		RowsTotal:    len(records),
		RowsInserted: len(records),
		RowsFailed:   0,
		Errors:       []ValidationError{},
	}, nil
}

func (uc *useCase) importRevenue(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error) {
	mineralMap, err := uc.repo.GetMineralCodeMap(ctx)
	if err != nil {
		return nil, err
	}

	records, validationErrors := parseRevenueCSV(req.File, req.CompanyID, userID, req.DataType, req.Version, req.Description, mineralMap)

	if len(validationErrors) > 0 {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    len(records) + len(validationErrors),
			RowsInserted: 0,
			RowsFailed:   len(validationErrors),
			Errors:       validationErrors,
		}, nil
	}

	err = uc.repo.InsertRevenueBulk(ctx, records)
	if err != nil {
		return nil, err
	}

	return &ImportResponse{
		Success:      true,
		Type:         req.Type,
		RowsTotal:    len(records),
		RowsInserted: len(records),
		RowsFailed:   0,
		Errors:       []ValidationError{},
	}, nil
}

func (uc *useCase) importFinancial(ctx context.Context, req *ImportRequest, userID int64) (*ImportResponse, error) {
	records, validationErrors := parseFinancialCSV(req.File, req.CompanyID, userID, req.DataType, req.Version, req.Description)

	if len(validationErrors) > 0 {
		return &ImportResponse{
			Success:      false,
			Type:         req.Type,
			RowsTotal:    len(records) + len(validationErrors),
			RowsInserted: 0,
			RowsFailed:   len(validationErrors),
			Errors:       validationErrors,
		}, nil
	}

	err := uc.repo.InsertFinancialBulk(ctx, records)
	if err != nil {
		return nil, err
	}

	return &ImportResponse{
		Success:      true,
		Type:         req.Type,
		RowsTotal:    len(records),
		RowsInserted: len(records),
		RowsFailed:   0,
		Errors:       []ValidationError{},
	}, nil
}
