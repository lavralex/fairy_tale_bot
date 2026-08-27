package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lavralex/fairy_tale_bot/internal/models"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

type createFieldAdminRequest struct {
	Name     string                     `json:"name"`
	IsActive bool                       `json:"is_active"`
	Options  []createOptionAdminRequest `json:"options"`
}
type createOptionAdminRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}
type fieldAdminResponse struct {
	ID        int64                 `json:"id"`
	Name      string                `json:"name"`
	IsActive  bool                  `json:"is_active"`
	Options   []optionAdminResponse `json:"options"`
	CreatedAt time.Time             `json:"created_at"`
}
type optionAdminResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func makeOptionResponses(options []models.FieldOption) []optionAdminResponse {
	optionsResponses := make([]optionAdminResponse, 0, len(options))
	for _, option := range options {
		optionsResponses = append(
			optionsResponses,
			optionAdminResponse{
				ID:        option.ID,
				Name:      option.Name,
				SortOrder: option.SortOrder,
				IsActive:  option.IsActive,
				CreatedAt: option.CreatedAt,
			},
		)
	}
	return optionsResponses
}

func (s *Server) createFieldAdmin(c echo.Context) error {
	var req createFieldAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	field := models.Field{
		Name:     req.Name,
		IsActive: req.IsActive,
	}
	for _, option := range req.Options {
		field.Options = append(
			field.Options,
			models.FieldOption{
				Name:      option.Name,
				SortOrder: option.SortOrder,
				IsActive:  option.IsActive,
			},
		)
	}
	item, err := s.store.CreateField(c.Request().Context(), field)
	if errors.Is(err, storage.ErrIncorrectNumberOfOptions) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "the incorrect number of options: required is 1 and more",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "create field error"})
	}
	createdField := fieldAdminResponse{
		ID:        item.ID,
		Name:      item.Name,
		IsActive:  item.IsActive,
		Options:   makeOptionResponses(item.Options),
		CreatedAt: item.CreatedAt,
	}
	return c.JSON(http.StatusCreated, createdField)
}

func (s *Server) getFieldAdmin(c echo.Context) error {
	fieldID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	item, err := s.store.GetField(c.Request().Context(), fieldID)
	if errors.Is(err, storage.ErrFieldNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "field not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "get field error"})
	}
	field := fieldAdminResponse{
		ID:        item.ID,
		Name:      item.Name,
		IsActive:  item.IsActive,
		Options:   makeOptionResponses(item.Options),
		CreatedAt: item.CreatedAt,
	}
	return c.JSON(http.StatusOK, field)
}

func (s *Server) listFieldsAdmin(c echo.Context) error {
	items, err := s.store.GetFields(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "get fields error"})
	}
	fields := make([]fieldAdminResponse, 0, len(items))
	for _, item := range items {
		fields = append(
			fields,
			fieldAdminResponse{
				ID:        item.ID,
				Name:      item.Name,
				IsActive:  item.IsActive,
				Options:   makeOptionResponses(item.Options),
				CreatedAt: item.CreatedAt,
			},
		)
	}
	return c.JSON(http.StatusOK, fields)
}

func (s *Server) deleteFieldAdmin(c echo.Context) error {
	fieldID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	err = s.store.DeleteField(c.Request().Context(), fieldID)
	if errors.Is(err, storage.ErrFieldNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "field not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "delete field error"})
	}
	return c.NoContent(http.StatusNoContent)
}

type updateFieldAdminRequest struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func (s *Server) updateFieldAdmin(c echo.Context) error {
	fieldID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	var req updateFieldAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	field := models.Field{
		ID:       fieldID,
		Name:     req.Name,
		IsActive: req.IsActive,
	}
	item, err := s.store.UpdateField(c.Request().Context(), field)
	if errors.Is(err, storage.ErrFieldNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "field not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update field error"})
	}
	updatedField := fieldAdminResponse{
		ID:        item.ID,
		Name:      item.Name,
		IsActive:  item.IsActive,
		Options:   makeOptionResponses(item.Options),
		CreatedAt: item.CreatedAt,
	}
	return c.JSON(http.StatusOK, updatedField)
}
