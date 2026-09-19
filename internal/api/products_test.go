package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/lavralex/fairy_tale_bot/internal/api/mocks"
	"github.com/lavralex/fairy_tale_bot/internal/models"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func makeTestProduct(id int64, templateID *int64) models.Product {
	return models.Product{
		ID:              id,
		Name:            "name",
		Description:     "description",
		Price:           "777",
		TemplateID:      templateID,
		Tags:            []string{"tag"},
		CreatedAt:       testDate,
		IsAvailable:     true,
		CommentEnabled:  true,
		TemplateEnabled: true,
		Images:          []models.ProductImage{makeTestImage(1, id)},
	}
}

func makeTestProductResponse(id int64, templateID *int64) *productAdminResponse {
	return &productAdminResponse{
		ID:              id,
		Name:            "name",
		Description:     "description",
		Price:           "777",
		Tags:            []string{"tag"},
		TemplateID:      templateID,
		IsAvailable:     true,
		CommentEnabled:  true,
		TemplateEnabled: true,
		CreatedAt:       testDate,
		Images:          []imageAdminResponse{*makeTestImageResponse(1)},
	}
}

func makeTestImage(id int64, ProductID int64) models.ProductImage {
	return models.ProductImage{
		ID:        id,
		ProductID: id,
		Src:       "src",
		Alt:       "alt",
		SortOrder: 1,
		CreatedAt: testDate,
	}
}

func makeTestImageResponse(id int64) *imageAdminResponse {
	return &imageAdminResponse{
		ID:        id,
		Src:       "src",
		Alt:       "alt",
		SortOrder: 1,
		CreatedAt: testDate,
	}
}

func TestCreateProductAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		body       string
		wantStatus int
		wantBody   *productAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateProduct(mock.Anything, models.Product{
					Name:            "name",
					Description:     "description",
					Price:           "777",
					Tags:            []string{"tag"},
					IsAvailable:     true,
					CommentEnabled:  true,
					TemplateEnabled: true,
					Images: []models.ProductImage{
						{Src: "src", Alt: "alt", SortOrder: 0},
					},
				}).Return(makeTestProduct(1, nil), nil)
			},
			body:       `{"name":"name","description":"description","price":"777","tags":["tag"],"is_available":true,"comment_enabled":true,"template_enabled":true,"images":[{"src":"src","alt":"alt"}]}`,
			wantStatus: http.StatusCreated,
			wantBody:   makeTestProductResponse(1, nil),
		},
		{
			name:       "invalid body",
			setupMock:  func(m *mocks.MockStorage) {},
			body:       `invalid json :(`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "incorrect number of images",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateProduct(mock.Anything, models.Product{
					Name:            "name",
					Description:     "description",
					Price:           "777",
					Tags:            []string{"tag"},
					IsAvailable:     true,
					CommentEnabled:  true,
					TemplateEnabled: true,
					Images: []models.ProductImage{
						{Src: "src", Alt: "alt", SortOrder: 0},
					},
				}).Return(models.Product{}, storage.ErrIncorrectNumberOfImages)
			},
			body:       `{"name":"name","description":"description","price":"777","tags":["tag"],"is_available":true,"comment_enabled":true,"template_enabled":true,"images":[{"src":"src","alt":"alt"}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateProduct(mock.Anything, models.Product{
					Name:            "name",
					Description:     "description",
					Price:           "777",
					Tags:            []string{"tag"},
					IsAvailable:     true,
					CommentEnabled:  true,
					TemplateEnabled: true,
					Images: []models.ProductImage{
						{Src: "src", Alt: "alt", SortOrder: 0},
					},
				}).Return(models.Product{}, errors.New("some storage error"))
			},
			body:       `{"name":"name","description":"description","price":"777","tags":["tag"],"is_available":true,"comment_enabled":true,"template_enabled":true,"images":[{"src":"src","alt":"alt"}]}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodPost, "/admin/products/", strings.NewReader(tt.body))
				gotErr := makeTestServer(t, tt.setupMock).createProductAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got productAdminResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
						t.Fatalf("unmarshal response: %v", err)
					}
					assert.Equal(t, *tt.wantBody, got)
				}
				assert.Equal(t, tt.wantStatus, rec.Code)
			},
		)
	}
}

func TestGetProductAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		id         string
		wantStatus int
		wantBody   *productAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetProduct(mock.Anything, int64(1)).Return(makeTestProduct(1, nil), nil)
			},
			id:         "1",
			wantStatus: http.StatusOK,
			wantBody:   makeTestProductResponse(1, nil),
		},
		{
			name:       "id can't parse",
			setupMock:  func(m *mocks.MockStorage) {},
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "not found",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetProduct(mock.Anything, int64(1)).Return(models.Product{}, storage.ErrProductNotFound)
			},
			id:         "1",
			wantStatus: http.StatusNotFound,
			wantBody:   nil,
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetProduct(mock.Anything, int64(1)).Return(models.Product{}, errors.New("some storage error"))
			},
			id:         "1",
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodGet, "/admin/products/"+tt.id, nil)
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := makeTestServer(t, tt.setupMock).getProductAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got productAdminResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
						t.Fatalf("unmarshal response: %v", err)
					}
					assert.Equal(t, *tt.wantBody, got)
				}
				assert.Equal(t, tt.wantStatus, rec.Code)
			},
		)
	}
}

func TestListProductAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		wantStatus int
		wantBody   []listProductsResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetProducts(mock.Anything, productsLimit, defaultOffset).
					Return([]models.Product{makeTestProduct(1, nil), makeTestProduct(2, nil)}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: []listProductsResponse{
				{
					ID: 1, Name: "name", Description: "description", Price: "777",
					Tags: []string{"tag"},
					Images: []listProductImageResponse{
						{ID: 1, Src: "src", Alt: "alt", SortOrder: 1},
					},
				},
				{
					ID: 2, Name: "name", Description: "description", Price: "777",
					Tags: []string{"tag"},
					Images: []listProductImageResponse{
						{ID: 1, Src: "src", Alt: "alt", SortOrder: 1},
					},
				},
			},
		},
		{
			name: "empty list",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetProducts(mock.Anything, productsLimit, defaultOffset).Return([]models.Product{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   []listProductsResponse{},
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetProducts(mock.Anything, productsLimit, defaultOffset).Return(nil, errors.New("some storage error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodGet, "/admin/products/", nil)
				gotErr := makeTestServer(t, tt.setupMock).listProducts(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got []listProductsResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
						t.Fatalf("unmarshal response: %v", err)
					}
					assert.Equal(t, tt.wantBody, got)
				}
				assert.Equal(t, tt.wantStatus, rec.Code)
			},
		)
	}
}

func TestDeleteProductAdmin(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupMock  func(m *mocks.MockStorage)
		wantStatus int
	}{
		{
			name: "valid value",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().DeleteProduct(mock.Anything, int64(1)).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "id can't parse",
			id:         "invalid",
			setupMock:  func(m *mocks.MockStorage) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().DeleteProduct(mock.Anything, int64(1)).Return(storage.ErrProductNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "storage error",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().DeleteProduct(mock.Anything, int64(1)).Return(errors.New("some storage error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := makeTestContext(http.MethodDelete, "/admin/products/"+tt.id, nil)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)
			gotErr := makeTestServer(t, tt.setupMock).deleteProductAdmin(c)
			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestUpdateProductAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		id         string
		body       string
		wantStatus int
		wantBody   *productAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateProduct(mock.Anything, models.Product{
					ID:              1,
					Name:            "name",
					Description:     "description",
					Price:           "777",
					Tags:            []string{"tag"},
					IsAvailable:     true,
					CommentEnabled:  true,
					TemplateEnabled: true,
					Images: []models.ProductImage{
						{Src: "src", Alt: "alt", SortOrder: 0},
					},
				}).Return(makeTestProduct(1, nil), nil)
			},
			id:         "1",
			body:       `{"name":"name","description":"description","price":"777","tags":["tag"],"is_available":true,"comment_enabled":true,"template_enabled":true,"images":[{"src":"src","alt":"alt"}]}`,
			wantStatus: http.StatusOK,
			wantBody:   makeTestProductResponse(1, nil),
		},
		{
			name:       "id can't parse",
			setupMock:  func(m *mocks.MockStorage) {},
			id:         "invalid",
			body:       `{"name":"name","description":"description","price":"777","tags":["tag"],"is_available":true,"comment_enabled":true,"template_enabled":true,"images":[{"src":"src","alt":"alt"}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "not found",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateProduct(mock.Anything, models.Product{
					ID:              1,
					Name:            "name",
					Description:     "description",
					Price:           "777",
					Tags:            []string{"tag"},
					IsAvailable:     true,
					CommentEnabled:  true,
					TemplateEnabled: true,
					Images: []models.ProductImage{
						{Src: "src", Alt: "alt", SortOrder: 0},
					},
				}).Return(models.Product{}, storage.ErrProductNotFound)
			},
			id:         "1",
			body:       `{"name":"name","description":"description","price":"777","tags":["tag"],"is_available":true,"comment_enabled":true,"template_enabled":true,"images":[{"src":"src","alt":"alt"}]}`,
			wantStatus: http.StatusNotFound,
			wantBody:   nil,
		},
		{
			name:       "invalid body",
			setupMock:  func(m *mocks.MockStorage) {},
			body:       `invalid json :(`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "incorrect number of images",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateProduct(mock.Anything, models.Product{
					ID:              1,
					Name:            "name",
					Description:     "description",
					Price:           "777",
					Tags:            []string{"tag"},
					IsAvailable:     true,
					CommentEnabled:  true,
					TemplateEnabled: true,
					Images: []models.ProductImage{
						{Src: "src", Alt: "alt", SortOrder: 0},
					},
				}).Return(models.Product{}, storage.ErrIncorrectNumberOfImages)
			},
			id:         "1",
			body:       `{"name":"name","description":"description","price":"777","tags":["tag"],"is_available":true,"comment_enabled":true,"template_enabled":true,"images":[{"src":"src","alt":"alt"}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateProduct(mock.Anything, models.Product{
					ID:              1,
					Name:            "name",
					Description:     "description",
					Price:           "777",
					Tags:            []string{"tag"},
					IsAvailable:     true,
					CommentEnabled:  true,
					TemplateEnabled: true,
					Images: []models.ProductImage{
						{Src: "src", Alt: "alt", SortOrder: 0},
					},
				}).Return(models.Product{}, errors.New("some storage error"))
			},
			id:         "1",
			body:       `{"name":"name","description":"description","price":"777","tags":["tag"],"is_available":true,"comment_enabled":true,"template_enabled":true,"images":[{"src":"src","alt":"alt"}]}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodPut, "/admin/products/"+tt.id, strings.NewReader(tt.body))
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := makeTestServer(t, tt.setupMock).updateProductAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got productAdminResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
						t.Fatalf("unmarshal response: %v", err)
					}
					assert.Equal(t, *tt.wantBody, got)
				}
				assert.Equal(t, tt.wantStatus, rec.Code)
			},
		)
	}
}
