package catalog

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/mytheresa/go-hiring-challenge/models"
)

func TestHandleGetByCode(t *testing.T) {
	tests := []struct {
		name           string
		productCode    string
		mockGetByCode  func(ctx context.Context, code string) (*models.Product, error)
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:        "success - variant inherits price from product",
			productCode: "P1",
			mockGetByCode: func(ctx context.Context, code string) (*models.Product, error) {
				return &models.Product{
					Code:     "P1",
					Price:    decimal.NewFromFloat(10),
					Category: models.Category{Code: "Code1", Name: "Name1"},
					Variants: []models.Variant{
						{Name: "Variant 1", SKU: "SKU001", Price: decimal.Zero},
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"code":"P1","price":10,"category":{"code":"Code1","name":"Name1"},"variants":[{"name":"Variant 1","sku":"SKU001","price":10}]}`,
		},
		{
			name:        "success - variant keeps its own specific price",
			productCode: "P2",
			mockGetByCode: func(ctx context.Context, code string) (*models.Product, error) {
				return &models.Product{
					Code:     "P2",
					Price:    decimal.NewFromFloat(50),
					Category: models.Category{Code: "Code2", Name: "Name2"},
					Variants: []models.Variant{
						{Name: "Variant Premium", SKU: "SKU999", Price: decimal.NewFromFloat(75)},
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"code":"P2","price":50,"category":{"code":"Code2","name":"Name2"},"variants":[{"name":"Variant Premium","sku":"SKU999","price":75}]}`,
		},
		{
			name:        "error - product not found",
			productCode: "UNKNOWN",
			mockGetByCode: func(ctx context.Context, code string) (*models.Product, error) {
				return nil, errors.New("product not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedJSON:   `{"error":"product not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepo{
				getByCodeFunc: tt.mockGetByCode,
			}

			h := NewCatalogHandler(repo)

			req := httptest.NewRequest(http.MethodGet, "/catalog/"+tt.productCode, nil)
			req.SetPathValue("code", tt.productCode)
			w := httptest.NewRecorder()

			h.HandleGetByCode(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, w.Body.String())
			}
		})
	}
}

func TestHandleGetAll(t *testing.T) {
	tests := []struct {
		name           string
		urlQuery       string
		mockExpect     func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error)
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:     "success - default pagination values (limit 10, offset 0) and no filters",
			urlQuery: "",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				assert.Equal(t, 10, limit)
				assert.Equal(t, 0, offset)
				assert.Empty(t, category)
				assert.Nil(t, priceLessThan)
				return 1, []models.Product{
					{
						Code:     "P1",
						Price:    decimal.NewFromFloat(25.50),
						Category: models.Category{Name: "Bags"},
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"total":1,"products":[{"code":"P1","price":25.5,"category":"Bags"}]}`,
		},
		{
			name:     "success - with valid category filter",
			urlQuery: "?category=boots",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				assert.Equal(t, "boots", category)
				assert.Nil(t, priceLessThan)
				return 0, []models.Product{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"total":0,"products":[]}`,
		},
		{
			name:     "success - with valid positive price_less_than filter",
			urlQuery: "?price_less_than=50.5",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				assert.NotNil(t, priceLessThan)
				assert.Equal(t, 50.5, *priceLessThan)
				assert.Empty(t, category)
				return 0, []models.Product{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"total":0,"products":[]}`,
		},
		{
			name:     "success - negative price_less_than filter is ignored",
			urlQuery: "?price_less_than=-10",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				assert.Nil(t, priceLessThan)
				return 0, []models.Product{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"total":0,"products":[]}`,
		},
		{
			name:     "success - non-numeric price_less_than filter is ignored",
			urlQuery: "?price_less_than=abc",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				assert.Nil(t, priceLessThan)
				return 0, []models.Product{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"total":0,"products":[]}`,
		},
		{
			name:     "success - with all filters combined simultaneously",
			urlQuery: "?limit=5&offset=15&category=sneakers&price_less_than=100",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				assert.Equal(t, 5, limit)
				assert.Equal(t, 15, offset)
				assert.Equal(t, "sneakers", category)
				assert.NotNil(t, priceLessThan)
				assert.Equal(t, 100.0, *priceLessThan)
				return 0, []models.Product{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"total":0,"products":[]}`,
		},
		{
			name:     "success - enforces minimum limit of 1 when limit < 1",
			urlQuery: "?limit=0",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				assert.Equal(t, 1, limit)
				return 0, []models.Product{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"total":0,"products":[]}`,
		},
		{
			name:     "success - enforces maximum limit of 100 when limit > 100",
			urlQuery: "?limit=250",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				assert.Equal(t, 100, limit)
				return 0, []models.Product{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"total":0,"products":[]}`,
		},
		{
			name:     "error - repository failure returns internal server error",
			urlQuery: "",
			mockExpect: func(ctx context.Context, limit, offset int, category string, priceLessThan *float64) (int64, []models.Product, error) {
				return 0, nil, errors.New("database connection failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedJSON:   `{"error":"database connection failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepo{
				getPaginatedFunc: tt.mockExpect,
			}

			h := NewCatalogHandler(repo)

			req := httptest.NewRequest(http.MethodGet, "/catalog"+tt.urlQuery, nil)
			w := httptest.NewRecorder()

			h.HandleGetAll(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, w.Body.String())
			}
		})
	}
}

type mockProductRepo struct {
	getPaginatedFunc func(ctx context.Context, offset, limit int, category string, priceLessThan *float64) (int64, []models.Product, error)
	getByCodeFunc    func(ctx context.Context, code string) (*models.Product, error)
}

func (s *mockProductRepo) GetPaginated(ctx context.Context, offset, limit int, category string, priceLessThan *float64) (int64, []models.Product, error) {
	return s.getPaginatedFunc(ctx, offset, limit, category, priceLessThan)
}

func (s *mockProductRepo) GetByCode(ctx context.Context, code string) (*models.Product, error) {
	return s.getByCodeFunc(ctx, code)
}
