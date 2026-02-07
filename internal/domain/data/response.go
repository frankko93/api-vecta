package data

// ImportResponse represents the response after importing data
type ImportResponse struct {
	Success      bool              `json:"success"`
	Type         DataImportType    `json:"type"`
	RowsTotal    int               `json:"rows_total"`
	RowsInserted int               `json:"rows_inserted"`
	RowsFailed   int               `json:"rows_failed"`
	Errors       []ValidationError `json:"errors,omitempty"`
}

// MessageResponse simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
