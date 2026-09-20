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

func makeTestTemplate(id int64) models.Template {
	return models.Template{
		ID:        id,
		Name:      "name",
		CreatedAt: testDate,
		Fields:    []models.TemplateField{makeTestTemplateField(1)},
	}
}

func makeTestTemplateResponse(id int64) *templateAdminResponse {
	return &templateAdminResponse{
		ID:        id,
		Name:      "name",
		CreatedAt: testDate,
		Fields:    []templateFieldAdminResponse{*makeTestTemplateFieldResponse(1)},
	}
}

func makeTestTemplateField(id int64) models.TemplateField {
	return models.TemplateField{
		Field:     makeTestField(id),
		SortOrder: 1,
	}
}

func makeTestTemplateFieldResponse(id int64) *templateFieldAdminResponse {
	return &templateFieldAdminResponse{
		Field:     *makeTestFieldResponse(id),
		SortOrder: 1,
	}
}

func TestCreateTemplateAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		body       string
		wantStatus int
		wantBody   *templateAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateTemplate(mock.Anything, models.Template{
					Name: "name",
					Fields: []models.TemplateField{
						{Field: models.Field{ID: 1}, SortOrder: 0},
					},
				}).Return(makeTestTemplate(1), nil)
			},
			body:       `{"name":"name","fields":[{"id":1}]}`,
			wantStatus: http.StatusCreated,
			wantBody:   makeTestTemplateResponse(1),
		},
		{
			name:       "invalid body",
			setupMock:  func(m *mocks.MockStorage) {},
			body:       `invalid json :(`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "incorrect number of fields",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateTemplate(mock.Anything, models.Template{
					Name: "name",
					Fields: []models.TemplateField{
						{Field: models.Field{ID: 1}, SortOrder: 0},
					},
				}).Return(models.Template{}, storage.ErrIncorrectNumberOfFields)
			},
			body:       `{"name":"name","fields":[{"id":1}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateTemplate(mock.Anything, models.Template{
					Name: "name",
					Fields: []models.TemplateField{
						{Field: models.Field{ID: 1}, SortOrder: 0},
					},
				}).Return(models.Template{}, errors.New("some storage error"))
			},
			body:       `{"name":"name","fields":[{"id":1}]}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodPost, "/admin/templates/", strings.NewReader(tt.body))
				gotErr := makeTestServer(t, tt.setupMock).createTemplateAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got templateAdminResponse
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

func TestGetTemplateAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		id         string
		wantStatus int
		wantBody   *templateAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetTemplate(mock.Anything, int64(1)).Return(makeTestTemplate(1), nil)
			},
			id:         "1",
			wantStatus: http.StatusOK,
			wantBody:   makeTestTemplateResponse(1),
		},
		{
			name:       "id parameter can't parse",
			setupMock:  func(m *mocks.MockStorage) {},
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "not found",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetTemplate(mock.Anything, int64(1)).Return(models.Template{}, storage.ErrTemplateNotFound)
			},
			id:         "1",
			wantStatus: http.StatusNotFound,
			wantBody:   nil,
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetTemplate(mock.Anything, int64(1)).Return(models.Template{}, errors.New("some storage error"))
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
				c, rec := makeTestContext(http.MethodGet, "/admin/templates/"+tt.id, nil)
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := makeTestServer(t, tt.setupMock).getTemplateAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got templateAdminResponse
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

func TestListTemplateAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		wantStatus int
		wantBody   []listTemplateResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetTemplates(mock.Anything, templatesLimit, defaultOffset).
					Return([]models.Template{makeTestTemplate(1), makeTestTemplate(2)}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: []listTemplateResponse{
				{ID: 1, Name: "name", CreatedAt: testDate},
				{ID: 2, Name: "name", CreatedAt: testDate},
			},
		},
		{
			name: "empty list",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetTemplates(mock.Anything, templatesLimit, defaultOffset).Return([]models.Template{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   []listTemplateResponse{},
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().GetTemplates(mock.Anything, templatesLimit, defaultOffset).Return(nil, errors.New("some storage error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodGet, "/admin/templates/", nil)
				gotErr := makeTestServer(t, tt.setupMock).listTemplatesAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got []listTemplateResponse
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

func TestDeleteTemplateAdmin(t *testing.T) {
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
				m.EXPECT().DeleteTemplate(mock.Anything, int64(1)).Return(nil)
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
				m.EXPECT().DeleteTemplate(mock.Anything, int64(1)).Return(storage.ErrTemplateNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "storage error",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().DeleteTemplate(mock.Anything, int64(1)).Return(errors.New("some storage error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := makeTestContext(http.MethodDelete, "/admin/templates/"+tt.id, nil)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)
			gotErr := makeTestServer(t, tt.setupMock).deleteTemplateAdmin(c)
			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestUpdateTemplateAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		id         string
		body       string
		wantStatus int
		wantBody   *templateAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateTemplate(mock.Anything, models.Template{
					ID:   1,
					Name: "name",
					Fields: []models.TemplateField{
						{Field: models.Field{ID: 1}, SortOrder: 0},
					},
				}).Return(makeTestTemplate(1), nil)
			},
			id:         "1",
			body:       `{"name":"name","fields":[{"id":1}]}`,
			wantStatus: http.StatusOK,
			wantBody:   makeTestTemplateResponse(1),
		},
		{
			name:       "id can't parse",
			setupMock:  func(m *mocks.MockStorage) {},
			id:         "invalid",
			body:       `{"name":"name","fields":[{"id":1}]}`,
			wantStatus: http.StatusBadRequest,
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
		{
			name: "not found",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateTemplate(mock.Anything, models.Template{
					ID:   1,
					Name: "name",
					Fields: []models.TemplateField{
						{Field: models.Field{ID: 1}, SortOrder: 0},
					},
				}).Return(models.Template{}, storage.ErrTemplateNotFound)
			},
			id:         "1",
			body:       `{"name":"name","fields":[{"id":1}]}`,
			wantStatus: http.StatusNotFound,
			wantBody:   nil,
		},
		{
			name: "incorrect number of fields",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateTemplate(mock.Anything, models.Template{
					ID:   1,
					Name: "name",
					Fields: []models.TemplateField{
						{Field: models.Field{ID: 1}, SortOrder: 0},
					},
				}).Return(models.Template{}, storage.ErrIncorrectNumberOfFields)
			},
			id:         "1",
			body:       `{"name":"name","fields":[{"id":1}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "storage error",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateTemplate(mock.Anything, models.Template{
					ID:   1,
					Name: "name",
					Fields: []models.TemplateField{
						{Field: models.Field{ID: 1}, SortOrder: 0},
					},
				}).Return(models.Template{}, errors.New("some storage error"))
			},
			id:         "1",
			body:       `{"name":"name","fields":[{"id":1}]}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodPut, "/admin/templates/"+tt.id, strings.NewReader(tt.body))
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := makeTestServer(t, tt.setupMock).updateTemplateAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got templateAdminResponse
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
