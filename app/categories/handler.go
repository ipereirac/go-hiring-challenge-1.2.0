package categories

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CategoriesHandler struct {
	repo models.CategoriesRepository
}

func NewCategoriesHandler(r models.CategoriesRepository) *CategoriesHandler {
	return &CategoriesHandler{
		repo: r,
	}
}

func (h *CategoriesHandler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetAll(r.Context())
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]categoryResponse, len(categories))
	for i, c := range categories {
		resp[i] = categoryResponse{Code: c.Code, Name: c.Name}
	}

	api.OKResponse(w, resp)
}

func (h *CategoriesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req categoryResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Body invalid.")
		return
	}

	if strings.TrimSpace(req.Code) == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Code required.")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Name required.")
		return
	}

	cat := models.Category{Code: req.Code, Name: req.Name}
	if err := h.repo.Create(r.Context(), &cat); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.OKResponse(w, categoryResponse{Code: cat.Code, Name: cat.Name})
}

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
