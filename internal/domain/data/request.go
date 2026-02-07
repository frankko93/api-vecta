package data

// ImportRequest represents a data import request
type ImportRequest struct {
	Type        DataImportType `form:"type" validate:"required"`
	DataType    string         `form:"data_type" validate:"required,oneof=actual budget"`
	CompanyID   int64          `form:"company_id" validate:"required,gt=0"`
	Version     int            `form:"version"`     // Optional, defaults to 1
	Description string         `form:"description"` // Optional
	File        []byte         `form:"-"`           // File content
}
