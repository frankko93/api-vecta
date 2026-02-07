package minerals

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
	minerals, err := h.useCase.List(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, minerals)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid mineral ID"))
		return
	}

	mineral, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrMineralNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, mineral)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req config.CreateMineralRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	mineral, err := h.useCase.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrCodeExists) {
			respond.Error(w, http.StatusConflict, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusCreated, mineral)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid mineral ID"))
		return
	}

	var req config.UpdateMineralRequest

	if err := request.DecodeJSON(w, r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err)
		return
	}

	mineral, err := h.useCase.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, ErrMineralNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, mineral)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, errors.New("invalid mineral ID"))
		return
	}

	err = h.useCase.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrMineralNotFound) {
			respond.Error(w, http.StatusNotFound, err)
			return
		}
		respond.Error(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, config.MessageResponse{Message: "mineral deleted successfully"})
}
