package categories

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mytheresa/go-hiring-challenge/models"
)

func TestHandleGetAll(t *testing.T) {
	tests := []struct {
		name           string
		mockGetAll     func(ctx context.Context) ([]models.Category, error)
		expectedStatus int
		expectedJSON   string
	}{
		{
			name: "success - returns all categories",
			mockGetAll: func(ctx context.Context) ([]models.Category, error) {
				return []models.Category{
					{Code: "CAT001", Name: "Shoes"},
					{Code: "CAT002", Name: "Bags"},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `[{"code":"CAT001","name":"Shoes"},{"code":"CAT002","name":"Bags"}]`,
		},
		{
			name: "success - empty list from repository",
			mockGetAll: func(ctx context.Context) ([]models.Category, error) {
				return []models.Category{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `[]`,
		},
		{
			name: "error - repository failure returns 500 internal server error",
			mockGetAll: func(ctx context.Context) ([]models.Category, error) {
				return nil, errors.New("database connection refused")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedJSON:   `{"error":"database connection refused"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCategoriesRepo{
				getAllFunc: tt.mockGetAll,
			}

			h := NewCategoriesHandler(repo)

			req := httptest.NewRequest(http.MethodGet, "/categories", nil)
			w := httptest.NewRecorder()

			h.HandleGetAll(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, w.Body.String())
			}
		})
	}
}

func TestHandleCreate(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockCreate     func(ctx context.Context, cat *models.Category) error
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:        "success - category created successfully",
			requestBody: `{"code":"CAT003","name":"Accessories"}`,
			mockCreate: func(ctx context.Context, cat *models.Category) error {
				assert.Equal(t, "CAT003", cat.Code)
				assert.Equal(t, "Accessories", cat.Name)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"code":"CAT003","name":"Accessories"}`,
		},
		{
			name:           "error - invalid json payload",
			requestBody:    `{"code":"CAT003", name:`,
			mockCreate:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   `{"error":"Body invalid."}`,
		},
		{
			name:           "error - missing code field",
			requestBody:    `{"name":"Accessories"}`,
			mockCreate:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   `{"error":"Code required."}`,
		},
		{
			name:           "error - blank code string",
			requestBody:    `{"code":"   ","name":"Accessories"}`,
			mockCreate:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   `{"error":"Code required."}`,
		},
		{
			name:           "error - missing name field",
			requestBody:    `{"code":"CAT003"}`,
			mockCreate:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   `{"error":"Name required."}`,
		},
		{
			name:           "error - blank name string",
			requestBody:    `{"code":"CAT003","name":"\t \n "}`,
			mockCreate:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   `{"error":"Name required."}`,
		},
		{
			name:        "error - repository database execution failure",
			requestBody: `{"code":"CAT003","name":"Accessories"}`,
			mockCreate: func(ctx context.Context, cat *models.Category) error {
				return errors.New("unique constraint violation: code already exists")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedJSON:   `{"error":"unique constraint violation: code already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCategoriesRepo{
				createFunc: tt.mockCreate,
			}

			h := NewCategoriesHandler(repo)

			req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			h.HandleCreate(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, w.Body.String())
			}
		})
	}
}

type mockCategoriesRepo struct {
	getAllFunc func(ctx context.Context) ([]models.Category, error)
	createFunc func(ctx context.Context, c *models.Category) error
}

func (m *mockCategoriesRepo) GetAll(ctx context.Context) ([]models.Category, error) {
	return m.getAllFunc(ctx)
}

func (m *mockCategoriesRepo) Create(ctx context.Context, c *models.Category) error {
	return m.createFunc(ctx, c)
}
