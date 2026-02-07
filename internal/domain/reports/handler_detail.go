package reports

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/gmhafiz/go8/internal/utility/respond"
)

// DetailHandler handles detailed report endpoints
type DetailHandler struct {
	useCase   DetailUseCase
	validator *validator.Validate
}

func NewDetailHandler(useCase DetailUseCase, validator *validator.Validate) *DetailHandler {
	return &DetailHandler{
		useCase:   useCase,
		validator: validator,
	}
}

// GetPBRDetail returns detailed PBR report
// @Summary Get detailed PBR report
// @Description Returns detailed PBR report with monthly breakdown and variances
// @Tags Reports
// @Accept json
// @Produce json
// @Param company_id query int true "Company ID"
// @Param year query int true "Year"
// @Param months query string false "Months filter (e.g., '1,2,3')"
// @Success 200 {object} PBRDetailReport
// @Router /api/v1/reports/pbr [get]
func (h *DetailHandler) GetPBRDetail(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseDetailRequest(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	report, err := h.useCase.GetPBRDetail(r.Context(), req)
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

// GetDoreDetail returns detailed Dore report
// @Summary Get detailed Dore report
// @Description Returns detailed Dore report with monthly breakdown, metal calculations, and variances
// @Tags Reports
// @Accept json
// @Produce json
// @Param company_id query int true "Company ID"
// @Param year query int true "Year"
// @Param months query string false "Months filter (e.g., '1,2,3')"
// @Success 200 {object} DoreDetailReport
// @Router /api/v1/reports/dore [get]
func (h *DetailHandler) GetDoreDetail(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseDetailRequest(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	report, err := h.useCase.GetDoreDetail(r.Context(), req)
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

// GetOPEXDetail returns detailed OPEX report
// @Summary Get detailed OPEX report
// @Description Returns detailed OPEX report with monthly breakdown, cost center and subcategory aggregations
// @Tags Reports
// @Accept json
// @Produce json
// @Param company_id query int true "Company ID"
// @Param year query int true "Year"
// @Param months query string false "Months filter (e.g., '1,2,3')"
// @Success 200 {object} OPEXDetailReport
// @Router /api/v1/reports/opex [get]
func (h *DetailHandler) GetOPEXDetail(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseDetailRequest(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	report, err := h.useCase.GetOPEXDetail(r.Context(), req)
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

// GetCAPEXDetail returns detailed CAPEX report
// @Summary Get detailed CAPEX report
// @Description Returns detailed CAPEX report with monthly breakdown, type and category aggregations
// @Tags Reports
// @Accept json
// @Produce json
// @Param company_id query int true "Company ID"
// @Param year query int true "Year"
// @Param months query string false "Months filter (e.g., '1,2,3')"
// @Success 200 {object} CAPEXDetailReport
// @Router /api/v1/reports/capex [get]
func (h *DetailHandler) GetCAPEXDetail(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseDetailRequest(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	report, err := h.useCase.GetCAPEXDetail(r.Context(), req)
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

// GetFinancialDetail returns detailed Financial report
// @Summary Get detailed Financial report
// @Description Returns detailed Financial report with monthly breakdown and variances
// @Tags Reports
// @Accept json
// @Produce json
// @Param company_id query int true "Company ID"
// @Param year query int true "Year"
// @Param months query string false "Months filter (e.g., '1,2,3')"
// @Success 200 {object} FinancialDetailReport
// @Router /api/v1/reports/financial [get]
func (h *DetailHandler) GetFinancialDetail(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseDetailRequest(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	report, err := h.useCase.GetFinancialDetail(r.Context(), req)
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

// GetProductionDetail returns detailed Production report
// @Summary Get detailed Production report
// @Description Returns detailed Production report with monthly breakdown by mineral
// @Tags Reports
// @Accept json
// @Produce json
// @Param company_id query int true "Company ID"
// @Param year query int true "Year"
// @Param months query string false "Months filter (e.g., '1,2,3')"
// @Success 200 {object} ProductionDetailReport
// @Router /api/v1/reports/production [get]
func (h *DetailHandler) GetProductionDetail(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseDetailRequest(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	report, err := h.useCase.GetProductionDetail(r.Context(), req)
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

// GetRevenueDetail returns detailed Revenue report
// @Summary Get detailed Revenue report
// @Description Returns detailed Revenue report with monthly breakdown by mineral
// @Tags Reports
// @Accept json
// @Produce json
// @Param company_id query int true "Company ID"
// @Param year query int true "Year"
// @Param months query string false "Months filter (e.g., '1,2,3')"
// @Success 200 {object} RevenueDetailReport
// @Router /api/v1/reports/revenue [get]
func (h *DetailHandler) GetRevenueDetail(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseDetailRequest(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	report, err := h.useCase.GetRevenueDetail(r.Context(), req)
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

// parseDetailRequest parses detail request from query parameters
func (h *DetailHandler) parseDetailRequest(r *http.Request) (*DetailRequest, error) {
	companyIDStr := r.URL.Query().Get("company_id")
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil || companyID <= 0 {
		return nil, errors.New("invalid or missing company_id")
	}

	yearStr := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 {
		return nil, errors.New("invalid or missing year")
	}

	months := r.URL.Query().Get("months")

	return &DetailRequest{
		CompanyID: companyID,
		Year:      year,
		Months:    months,
	}, nil
}
