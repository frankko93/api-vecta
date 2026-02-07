package data

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/gmhafiz/go8/internal/middleware"
	"github.com/gmhafiz/go8/internal/utility/respond"
)

const (
	// MaxUploadSize is 10MB
	MaxUploadSize = 10 << 20
)

type Handler struct {
	useCase   UseCase
	validator *validator.Validate
}

// NewHandler creates a new data handler
func NewHandler(uc UseCase, validator *validator.Validate) *Handler {
	return &Handler{
		useCase:   uc,
		validator: validator,
	}
}

// RegisterHTTPEndPoints registers data import HTTP endpoints
// Deprecated: Use NewHandler and register routes in initDomains for role-based access control
func RegisterHTTPEndPoints(router *chi.Mux, validator *validator.Validate, uc UseCase) {
	h := NewHandler(uc, validator)

	router.Route("/api/v1/data", func(r chi.Router) {
		r.Post("/import", h.Import)
		r.Get("/{type}/list", h.List)
		r.Delete("/{type}/{id}", h.Delete)
	})
}

// Import handles CSV data import
// @Summary Import data from CSV
// @Description Import production, dore, pbr, opex, capex or revenue data from CSV
// @Tags data
// @Accept multipart/form-data
// @Produce json
// @Param type formData string true "Data type" Enums(production, dore, pbr, opex, capex, revenue)
// @Param company_id formData integer true "Company ID"
// @Param file formData file true "CSV file"
// @Success 200 {object} ImportResponse
// @Failure 400 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /api/v1/data/import [post]
func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	// Parse multipart form
	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("file too large or invalid form data"))
		return
	}

	// Get type
	importType := DataImportType(r.FormValue("type"))
	if !importType.IsValid() {
		respond.Error(w, http.StatusBadRequest, ErrInvalidDataType)
		return
	}

	// Get data_type (actual or budget)
	dataType := DataType(r.FormValue("data_type"))
	if !dataType.IsValid() {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid data_type: must be 'actual' or 'budget'"))
		return
	}

	// Get company ID
	companyIDStr := r.FormValue("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil || companyID <= 0 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company_id"))
		return
	}

	// Get file
	file, _, err := r.FormFile("file")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("missing or invalid file"))
		return
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, errors.New("error reading file"))
		return
	}

	// Create import request
	importReq := &ImportRequest{
		Type:      importType,
		DataType:  string(dataType),
		CompanyID: companyID,
		File:      fileContent,
	}

	// Process import
	response, err := h.useCase.ImportData(r.Context(), importReq, userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	// If validation failed, return 400 with errors
	if !response.Success {
		respond.JSON(w, http.StatusBadRequest, response)
		return
	}

	respond.JSON(w, http.StatusOK, response)
}

// List returns imported data
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	dataTypeStr := chi.URLParam(r, "type")
	dataType := DataImportType(dataTypeStr)
	if !dataType.IsValid() {
		respond.Error(w, http.StatusBadRequest, ErrInvalidDataType)
		return
	}

	companyIDStr := r.URL.Query().Get("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil || companyID <= 0 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid or missing company_id"))
		return
	}

	yearStr := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid or missing year"))
		return
	}

	typeFilter := r.URL.Query().Get("data_type")
	if typeFilter == "" {
		typeFilter = "actual"
	}

	versionStr := r.URL.Query().Get("version")
	version := 1
	if versionStr != "" {
		version, err = strconv.Atoi(versionStr)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, errors.New("invalid version"))
			return
		}
	}

	data, err := h.useCase.ListData(r.Context(), dataType, companyID, year, typeFilter, version)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, data)
}

// Delete soft deletes imported data
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	dataTypeStr := chi.URLParam(r, "type")
	dataType := DataImportType(dataTypeStr)
	if !dataType.IsValid() {
		respond.Error(w, http.StatusBadRequest, ErrInvalidDataType)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	err = h.useCase.DeleteData(r.Context(), dataType, id)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, MessageResponse{Message: "data deleted successfully"})
}
