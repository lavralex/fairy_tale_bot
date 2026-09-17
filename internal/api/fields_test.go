package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lavralex/fairy_tale_bot/internal/models"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
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

func TestCreateFieldAdmin(t *testing.T) {
	tests := []struct {
		name       string
		server     *Server
		body       string
		wantStatus int
		wantBody   *fieldAdminResponse
	}{
		{
			name: "valid value",
			server: &Server{
				store: &fakeStorage{
					createFieldFunc: func(ctx context.Context, field models.Field) (models.Field, error) {
						return makeTestField(1), nil
					},
				},
			},
			body:       `{"name":"Материал","is_active":true,"options":[{"name":"Хлопок"}]}`,
			wantStatus: http.StatusCreated,
			wantBody:   makeTestFieldResponse(1),
		},
		{
			name: "incorrect number of options",
			server: &Server{
				store: &fakeStorage{
					createFieldFunc: func(ctx context.Context, field models.Field) (models.Field, error) {
						return models.Field{}, storage.ErrIncorrectNumberOfOptions
					},
				},
			},
			body:       `{"name":"Материал","is_active":true,"options":[{"name":"Хлопок"}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "invalid body",
			server: &Server{
				store: &fakeStorage{
					createFieldFunc: func(ctx context.Context, field models.Field) (models.Field, error) {
						return models.Field{}, nil
					},
				},
			},
			body:       `invalid json :(`,
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/admin/fields/", strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				e := echo.New()
				c := e.NewContext(req, rec)
				gotErr := tt.server.createFieldAdmin(c)
				if gotErr != nil {
					t.Fatalf("createFieldAdmin() unexpected error = %v", gotErr)
				}
				if tt.wantBody != nil {
					var got fieldAdminResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
						t.Fatalf("unmarshal response: %v", err)
					}
					if !reflect.DeepEqual(got, *tt.wantBody) {
						t.Errorf("body = %+v, want %+v", got, *tt.wantBody)
					}
				}
				if rec.Code != tt.wantStatus {
					t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
				}
			},
		)
	}
}

func TestGetFieldAdmin(t *testing.T) {
	tests := []struct {
		name       string
		server     *Server
		id         string
		wantStatus int
		wantBody   *fieldAdminResponse
	}{
		{
			name: "valid value",
			server: &Server{
				store: &fakeStorage{
					getFieldFunc: func(ctx context.Context, id int64) (models.Field, error) {
						return makeTestField(1), nil
					},
				},
			},
			id:         "1",
			wantStatus: http.StatusOK,
			wantBody:   makeTestFieldResponse(1),
		},
		{
			name:       "invalid id value",
			server:     &Server{},
			id:         "",
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name: "field not found",
			server: &Server{
				store: &fakeStorage{
					getFieldFunc: func(ctx context.Context, id int64) (models.Field, error) {
						return models.Field{}, storage.ErrFieldNotFound
					},
				},
			},
			id:         "1",
			wantStatus: http.StatusNotFound,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/admin/fields/"+tt.id, nil)
				rec := httptest.NewRecorder()
				e := echo.New()
				c := e.NewContext(req, rec)
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := tt.server.getFieldAdmin(c)
				if gotErr != nil {
					t.Fatalf("getFieldAdmin() unexpected error = %v", gotErr)
				}
				if tt.wantBody != nil {
					var got fieldAdminResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
						t.Fatalf("unmarshal response: %v", err)
					}
					if !reflect.DeepEqual(got, *tt.wantBody) {
						t.Errorf("body = %+v, want %+v", got, *tt.wantBody)
					}
				}
				if rec.Code != tt.wantStatus {
					t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
				}
			},
		)
	}
}

func TestDeleteFieldAdmin(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		server     *Server
		wantStatus int
	}{
		{
			name: "valid value",
			id:   "1",
			server: &Server{
				store: &fakeStorage{
					deleteFieldFunc: func(ctx context.Context, id int64) error {
						return nil
					},
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "invalid id value",
			id:   "incorrect",
			server: &Server{
				store: &fakeStorage{
					deleteFieldFunc: func(ctx context.Context, id int64) error {
						return nil
					},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "1",
			server: &Server{
				store: &fakeStorage{
					deleteFieldFunc: func(ctx context.Context, id int64) error {
						return storage.ErrFieldNotFound
					},
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				req := httptest.NewRequest(http.MethodDelete, "/admin/fields/"+tt.id, nil)
				rec := httptest.NewRecorder()
				e := echo.New()
				c := e.NewContext(req, rec)
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
				gotErr := tt.server.deleteFieldAdmin(c)
				if gotErr != nil {
					t.Fatalf("deleteFieldAdmin() unexpected error = %v", gotErr)
				}
				if rec.Code != tt.wantStatus {
					t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
				}
			},
		)
	}
}
