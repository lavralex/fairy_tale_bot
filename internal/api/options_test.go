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

func TestCreateOptionAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		body       string
		wantStatus int
		wantBody   *optionAdminResponse
	}{
		{
			name: "valid value",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateOption(mock.Anything, models.FieldOption{
					Name:      "name",
					FieldID:   1,
					SortOrder: 1,
					IsActive:  true,
				}).Return(makeTestOption(1, 1), nil)
			},
			body:       `{"name":"name","field_id":1,"sort_order":1,"is_active":true}`,
			wantStatus: http.StatusCreated,
			wantBody:   makeTestOptionResponse(1),
		},
		{
			name: "field not found",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().CreateOption(mock.Anything, models.FieldOption{
					Name:      "name",
					FieldID:   1,
					SortOrder: 1,
					IsActive:  true,
				}).Return(models.FieldOption{}, storage.ErrFieldNotFound)
			},
			body:       `{"name":"name","field_id":1,"sort_order":1,"is_active":true}`,
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
				m.EXPECT().CreateOption(mock.Anything, models.FieldOption{
					Name:      "name",
					FieldID:   1,
					SortOrder: 1,
					IsActive:  true,
				}).Return(models.FieldOption{}, errors.New("some storage error"))
			},
			body:       `{"name":"name","field_id":1,"sort_order":1,"is_active":true}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodPost, "/admin/options/", strings.NewReader(tt.body))
				gotErr := makeTestServer(t, tt.setupMock).createOptionAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got optionAdminResponse
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

func TestDeleteOptionAdmin(t *testing.T) {
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
				m.EXPECT().DeleteOption(mock.Anything, int64(1)).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid id value",
			id:         "invalid",
			setupMock:  func(m *mocks.MockStorage) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().DeleteOption(mock.Anything, int64(1)).Return(storage.ErrOptionNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "storage error",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().DeleteOption(mock.Anything, int64(1)).Return(errors.New("some storage error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := makeTestContext(http.MethodDelete, "/admin/options/"+tt.id, nil)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)
			gotErr := makeTestServer(t, tt.setupMock).deleteOptionAdmin(c)
			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestUpdateOptionAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockStorage)
		id         string
		body       string
		wantStatus int
		wantBody   *optionAdminResponse
	}{
		{
			name: "valid value",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateOption(mock.Anything, models.FieldOption{
					ID:        1,
					Name:      "name",
					SortOrder: 1,
					IsActive:  true,
				}).Return(makeTestOption(1, 1), nil)
			},
			body:       `{"name":"name","sort_order":1,"is_active":true}`,
			wantStatus: http.StatusOK,
			wantBody:   makeTestOptionResponse(1),
		},
		{
			name:       "id parameter can't parse",
			id:         "invalid",
			setupMock:  func(m *mocks.MockStorage) {},
			body:       `{"name":"name","sort_order":1,"is_active":true}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name:       "invalid body",
			id:         "1",
			setupMock:  func(m *mocks.MockStorage) {},
			body:       `invalid json :(`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "not found",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateOption(mock.Anything, models.FieldOption{
					ID:        1,
					Name:      "name",
					SortOrder: 1,
					IsActive:  true,
				}).Return(models.FieldOption{}, storage.ErrOptionNotFound)
			},
			body:       `{"name":"name","sort_order":1,"is_active":true}`,
			wantStatus: http.StatusNotFound,
			wantBody:   nil,
		},
		{
			name: "storage error",
			id:   "1",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().UpdateOption(mock.Anything, models.FieldOption{
					ID:        1,
					Name:      "name",
					SortOrder: 1,
					IsActive:  true,
				}).Return(models.FieldOption{}, errors.New("some storage error"))
			},
			body:       `{"name":"name","sort_order":1,"is_active":true}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				c, rec := makeTestContext(http.MethodPut, "/admin/options/"+tt.id, strings.NewReader(tt.body))
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := makeTestServer(t, tt.setupMock).updateOptionAdmin(c)
				require.NoError(t, gotErr)
				if tt.wantBody != nil {
					var got optionAdminResponse
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
