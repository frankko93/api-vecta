package data

import "context"

// ListData returns imported data for a specific type, company, year and version
func (uc *useCase) ListData(ctx context.Context, dataType DataImportType, companyID int64, year int, typeFilter string, version int) (interface{}, error) {
	if version == 0 {
		version = 1
	}

	switch dataType {
	case ImportPBR:
		return uc.repo.ListPBRData(ctx, companyID, year, typeFilter, version)
	case ImportDore:
		return uc.repo.ListDoreData(ctx, companyID, year, typeFilter, version)
	case ImportOPEX:
		return uc.repo.ListOPEXData(ctx, companyID, year, typeFilter, version)
	case ImportCAPEX:
		return uc.repo.ListCAPEXData(ctx, companyID, year, typeFilter, version)
	case ImportFinancial:
		return uc.repo.ListFinancialData(ctx, companyID, year, typeFilter, version)
	default:
		return nil, ErrInvalidDataType
	}
}

// DeleteData soft deletes an imported data record
func (uc *useCase) DeleteData(ctx context.Context, dataType DataImportType, id int64) error {
	switch dataType {
	case ImportPBR:
		return uc.repo.SoftDeletePBRData(ctx, id)
	case ImportDore:
		return uc.repo.SoftDeleteDoreData(ctx, id)
	case ImportOPEX:
		return uc.repo.SoftDeleteOPEXData(ctx, id)
	case ImportCAPEX:
		return uc.repo.SoftDeleteCAPEXData(ctx, id)
	case ImportFinancial:
		return uc.repo.SoftDeleteFinancialData(ctx, id)
	default:
		return ErrInvalidDataType
	}
}
