package reports

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	authRepo "github.com/gmhafiz/go8/internal/domain/auth/repository"
	"github.com/gmhafiz/go8/internal/middleware"
	"github.com/gmhafiz/go8/internal/utility/request"
	"github.com/gmhafiz/go8/internal/utility/respond"
)

type Handler struct {
	useCase   UseCase
	validator *validator.Validate
	authRepo  authRepo.Repository
}

// NewHandler creates a new reports handler
func NewHandler(uc UseCase, validator *validator.Validate, authRepository authRepo.Repository) *Handler {
	return &Handler{
		useCase:   uc,
		validator: validator,
		authRepo:  authRepository,
	}
}

// RegisterHTTPEndPoints registers reports HTTP endpoints
// Deprecated: Use NewHandler and register routes in initDomains for role-based access control
func RegisterHTTPEndPoints(router *chi.Mux, validator *validator.Validate, uc UseCase, detailUC DetailUseCase, authRepository authRepo.Repository) {
	h := NewHandler(uc, validator, authRepository)
	detailH := NewDetailHandler(detailUC, validator, authRepository)

	router.Route("/api/v1/reports", func(r chi.Router) {
		r.Get("/summary", h.GetSummary)
		r.Post("/save", h.SaveReport)
		r.Get("/saved", h.ListSavedReports)
		r.Post("/compare", h.CompareReports)

		// Detailed reports
		r.Get("/pbr", detailH.GetPBRDetail)
		r.Get("/dore", detailH.GetDoreDetail)
		r.Get("/opex", detailH.GetOPEXDetail)
		r.Get("/capex", detailH.GetCAPEXDetail)
	})
}

// GetSummary returns the summary report for a company
// @Summary Get summary report
// @Description Get complete summary report with actual and budget data
// @Tags reports
// @Produce json
// @Param company_id query integer true "Company ID"
// @Param year query integer true "Year"
// @Param budget_version query integer true "Budget version to compare against"
// @Param months query string false "Comma-separated months (1-12)" example:"1,2,3"
// @Success 200 {object} SummaryReport
// @Failure 400 {object} respond.Error
// @Failure 404 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /api/v1/reports/summary [get]
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	// Parse company_id
	companyIDStr := r.URL.Query().Get("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil || companyID <= 0 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid or missing company_id"))
		return
	}

	// Parse year
	yearStr := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid or missing year"))
		return
	}

	// Parse budget_version (required)
	budgetVersionStr := r.URL.Query().Get("budget_version")
	budgetVersion, err := strconv.Atoi(budgetVersionStr)
	if err != nil || budgetVersion < 1 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid or missing budget_version (must be >= 1)"))
		return
	}

	// Parse months (optional - if empty, returns all 12 months)
	months := r.URL.Query().Get("months")

	req := &SummaryRequest{
		CompanyID:     companyID,
		Year:          year,
		Months:        months,
		BudgetVersion: budgetVersion,
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	report, err := h.useCase.GetSummary(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrCompanyNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, report)
}

// SaveReport saves a report snapshot
// Requires: editor role (can create/modify data)
func (h *Handler) SaveReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	var req SaveReportRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	// Validate user has editor role in this company (from session cache - no DB query)
	if err := middleware.CheckCompanyRole(r.Context(), req.CompanyID, middleware.RoleEditor); err != nil {
		if errors.Is(err, middleware.ErrCompanyAccessDenied) || errors.Is(err, middleware.ErrInsufficientRole) {
			respond.Error(w, http.StatusForbidden, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	savedReport, err := h.useCase.SaveReport(r.Context(), &req, userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusCreated, savedReport)
}

// ListSavedReports returns saved reports for a company and year
func (h *Handler) ListSavedReports(w http.ResponseWriter, r *http.Request) {
	companyIDStr := r.URL.Query().Get("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil || companyID <= 0 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company_id"))
		return
	}

	yearStr := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid year"))
		return
	}

	reports, err := h.useCase.ListSavedReports(r.Context(), companyID, year)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, reports)
}

// CompareReports compares multiple saved reports
// Requires: viewer role (read-only access is sufficient to compare)
func (h *Handler) CompareReports(w http.ResponseWriter, r *http.Request) {
	var req CompareReportsRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	// Validate user has at least viewer access to all reports' companies
	for _, reportID := range req.ReportIDs {
		companyID, err := h.useCase.GetReportCompanyID(r.Context(), reportID)
		if err != nil {
			if errors.Is(err, ErrReportNotFound) {
				respond.Error(w, http.StatusNotFound, err)
				return
			}
			respond.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Viewer role is sufficient to compare reports (from session cache - no DB query)
		if err := middleware.CheckCompanyRole(r.Context(), companyID, middleware.RoleViewer); err != nil {
			if errors.Is(err, middleware.ErrCompanyAccessDenied) || errors.Is(err, middleware.ErrInsufficientRole) {
				respond.Error(w, http.StatusForbidden, errors.New("you don't have access to one or more of the selected reports"))
				return
			}
			respond.Error(w, http.StatusInternalServerError, err)
			return
		}
	}

	comparison, err := h.useCase.CompareReports(r.Context(), req.ReportIDs)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, comparison)
}
