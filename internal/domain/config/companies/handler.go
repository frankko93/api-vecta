package companies

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/gmhafiz/go8/internal/domain/config"
	"github.com/gmhafiz/go8/internal/utility/request"
	"github.com/gmhafiz/go8/internal/utility/respond"
)

type Handler struct {
	useCase   UseCase
	validator *validator.Validate
}

func NewHandler(uc UseCase, validator *validator.Validate) *Handler {
	return &Handler{
		useCase:   uc,
		validator: validator,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	companies, err := h.useCase.List(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, companies)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company ID"))
		return
	}

	company, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrCompanyNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, company)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req config.CreateCompanyRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	company, err := h.useCase.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrTaxIDExists) {
			respond.Error(w, http.StatusConflict, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusCreated, company)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company ID"))
		return
	}

	var req config.UpdateCompanyRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	company, err := h.useCase.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, ErrCompanyNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, company)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company ID"))
		return
	}

	err = h.useCase.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrCompanyNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, config.MessageResponse{Message: "company deleted successfully"})
}

func (h *Handler) AssignMinerals(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company ID"))
		return
	}

	var req config.AssignMineralsRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.useCase.AssignMinerals(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, ErrCompanyNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, config.MessageResponse{Message: "minerals assigned successfully"})
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid company ID"))
		return
	}

	var req config.UpdateCompanySettingsRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	settings, err := h.useCase.UpdateSettings(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, ErrCompanyNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, settings)
}

func (h *Handler) GetAvailableUnits(w http.ResponseWriter, r *http.Request) {
	units := h.useCase.GetAvailableUnits(r.Context())
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": units})
}
