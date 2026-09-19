package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lavralex/fairy_tale_bot/internal/api/mocks"
	"github.com/lavralex/fairy_tale_bot/internal/models"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testDate = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func makeTestField(id int64) models.Field {
	return models.Field{
		ID:        id,
		Name:      "name",
		IsActive:  true,
		CreatedAt: testDate,
		Options: []models.FieldOption{
			makeTestOption(1, id),
		},
	}
}

func makeTestFieldResponse(id int64) *fieldAdminResponse {
	return &fieldAdminResponse{
		ID:        id,
		Name:      "name",
		IsActive:  true,
		CreatedAt: testDate,
		Options: []optionAdminResponse{
			*makeTestOptionResponse(1),
		},
	}
}

func makeTestOption(id int64, fieldID int64) models.FieldOption {
	return models.FieldOption{
		ID:        id,
		Name:      "name",
		FieldID:   fieldID,
		IsActive:  true,
		SortOrder: 1,
		CreatedAt: testDate,
	}
}

func makeTestOptionResponse(id int64) *optionAdminResponse {
	return &optionAdminResponse{
		ID:        id,
		Name:      "name",
		SortOrder: 1,
		IsActive:  true,
		CreatedAt: testDate,
	}
}

func makeTestContext(method string, path string, body io.Reader) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, body)
	rec := httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	e := echo.New()
	return e.NewContext(req, rec), rec
}

func makeTestServer(t *testing.T, setup func(m *mocks.MockStorage)) *Server {
	m := mocks.NewMockStorage(t)
	setup(m)
	return &Server{store: m}
}

func TestCreateFieldAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		body       string
		wantStatus int
		wantBody   *fieldAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateField(
					mock.Anything,
					models.Field{
						Name:     "name",
						IsActive: true,
						Options: []models.FieldOption{
							{Name: "name", IsActive: true, SortOrder: 0},
						},
					},
				).Return(makeTestField(1), nil)
			},
			body:       `{"name":"name","is_active":true,"options":[{"name":"name","is_active":true}]}`,
			wantStatus: http.StatusCreated,
			wantBody:   makeTestFieldResponse(1),
		},
		{
			name: "incorrect number of options",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateField(mock.Anything, models.Field{Name: "name"}).Return(models.Field{}, storage.ErrIncorrectNumberOfOptions)
			},
			body:       `{"name":"name"}`,
			wantStatus: http.StatusBadRequest,
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
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateField(mock.Anything, models.Field{
					Name:     "name",
					IsActive: true,
					Options: []models.FieldOption{
						{Name: "name", IsActive: true, SortOrder: 0},
					},
				}).Return(models.Field{}, errors.New("some storage error"))
			},
			body:       `{"name":"name","is_active":true,"options":[{"name":"name","is_active":true}]}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodPost, "/admin/fields/", strings.NewReader(tt.body))
				gotErr := makeTestServer(t, tt.setupMock).createFieldAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got fieldAdminResponse
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

func TestGetFieldAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		id         string
		wantStatus int
		wantBody   *fieldAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetField(mock.Anything, int64(1)).Return(makeTestField(1), nil)
			},
			id:         "1",
			wantStatus: http.StatusOK,
			wantBody:   makeTestFieldResponse(1),
		},
		{
			name:       "invalid id value",
			setupMock:  func(m *mocks.MockStorage) {},
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "field not found",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetField(mock.Anything, int64(1)).Return(models.Field{}, storage.ErrFieldNotFound)
			},
			id:         "1",
			wantStatus: http.StatusNotFound,
			wantBody:   nil,
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetField(mock.Anything, int64(1)).Return(models.Field{}, errors.New("some storage error"))
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
				c, rec := makeTestContext(http.MethodGet, "/admin/fields/"+tt.id, nil)
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := makeTestServer(t, tt.setupMock).getFieldAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got fieldAdminResponse
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

func TestListFieldsAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		wantStatus int
		wantBody   []fieldAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetFields(mock.Anything).Return([]models.Field{makeTestField(1), makeTestField(2)}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   []fieldAdminResponse{*makeTestFieldResponse(1), *makeTestFieldResponse(2)},
		},
		{
			name: "empty list",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetFields(mock.Anything).Return([]models.Field{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   []fieldAdminResponse{},
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetFields(mock.Anything).Return(nil, errors.New("some storage error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodGet, "/admin/fields/", nil)
				gotErr := makeTestServer(t, tt.setupMock).listFieldsAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got []fieldAdminResponse
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

func TestDeleteFieldAdmin(t *testing.T) {
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
				m.EXPECT().DeleteField(mock.Anything, int64(1)).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid id value",
			id:         "incorrect",
			setupMock:  func(m *mocks.MockStorage) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().DeleteField(mock.Anything, int64(1)).Return(storage.ErrFieldNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "storage error",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().DeleteField(mock.Anything, int64(1)).Return(errors.New("some storage error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := makeTestContext(http.MethodDelete, "/admin/fields/"+tt.id, nil)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)
			gotErr := makeTestServer(t, tt.setupMock).deleteFieldAdmin(c)
			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestUpdateFieldAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		id         string
		body       string
		wantStatus int
		wantBody   *fieldAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateField(mock.Anything, models.Field{
					ID:       1,
					Name:     "name",
					IsActive: true,
				}).Return(makeTestField(1), nil)
			},
			id:         "1",
			body:       `{"name":"name","is_active":true}`,
			wantStatus: http.StatusOK,
			wantBody:   makeTestFieldResponse(1),
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateField(mock.Anything, models.Field{
					ID:       1,
					Name:     "name",
					IsActive: true,
				}).Return(models.Field{}, errors.New("some storage error"))
			},
			id:         "1",
			body:       `{"name":"name","is_active":true}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
		{
			name:       "invalid id value",
			setupMock:  func(m *mocks.MockStorage) {},
			id:         "invalid",
			body:       `{"name":"name","is_active":true}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "field not found",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateField(mock.Anything, models.Field{
					ID:       1,
					Name:     "name",
					IsActive: true,
				}).Return(models.Field{}, storage.ErrFieldNotFound)
			},
			id:         "1",
			body:       `{"name":"name","is_active":true}`,
			wantStatus: http.StatusNotFound,
			wantBody:   nil,
		},
		{
			name:       "invalid body",
			setupMock:  func(m *mocks.MockStorage) {},
			id:         "1",
			body:       `invalid json :(`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodPut, "/admin/fields/"+tt.id, strings.NewReader(tt.body))
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := makeTestServer(t, tt.setupMock).updateFieldAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got fieldAdminResponse
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
