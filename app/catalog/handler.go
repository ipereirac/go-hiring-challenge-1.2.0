package catalog

import (
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type Response struct {
	Total    int64     `json:"total"`
	Products []Product `json:"products"`
}

type Product struct {
	Code     string  `json:"code"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type CatalogHandler struct {
	repo models.ProductsRepository
}

func NewCatalogHandler(r models.ProductsRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			if val < 1 {
				limit = 1
			} else if val > 100 {
				limit = 100
			} else {
				limit = val
			}
		}
	}

	categoryFilter := r.URL.Query().Get("category")

	var priceLessThanFilter *float64
	if priceStr := r.URL.Query().Get("price_less_than"); priceStr != "" {
		if val, err := strconv.ParseFloat(priceStr, 64); err == nil && val >= 0 {
			priceLessThanFilter = &val
		}
	}

	count, res, err := h.repo.GetPaginated(r.Context(), limit, offset, categoryFilter, priceLessThanFilter)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:     p.Code,
			Price:    p.Price.InexactFloat64(),
			Category: p.Category.Name,
		}
	}

	response := Response{
		Total:    count,
		Products: products,
	}

	api.OKResponse(w, response)
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	p, err := h.repo.GetByCode(r.Context(), code)
	if err != nil {
		api.ErrorResponse(w, http.StatusNotFound, "product not found")
		return
	}

	response := &productResponse{
		Code:  p.Code,
		Price: p.Price.InexactFloat64(),
		Category: categoryResponse{
			Code: p.Category.Code,
			Name: p.Category.Name,
		},
	}

	response.Variants = make([]variantResponse, len(p.Variants))
	for i, v := range p.Variants {
		price := v.Price.InexactFloat64()
		if price == 0 {
			price = p.Price.InexactFloat64()
		}
		response.Variants[i] = variantResponse{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price,
		}
	}
	api.OKResponse(w, response)
}

type productResponse struct {
	Code     string            `json:"code"`
	Price    float64           `json:"price"`
	Category categoryResponse  `json:"category"`
	Variants []variantResponse `json:"variants,omitempty"`
}

type variantResponse struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
