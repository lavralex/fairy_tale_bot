package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/lavralex/fairy_tale_bot/internal/models"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

func makeOptionResponse(option models.FieldOption) optionAdminResponse {
	return optionAdminResponse{
		ID:        option.ID,
		Name:      option.Name,
		SortOrder: option.SortOrder,
		IsActive:  option.IsActive,
		CreatedAt: option.CreatedAt,
	}
}

type createOptionAdminRequest struct {
	Name      string `json:"name"`
	FieldID   int64  `json:"field_id"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

func (s *Server) createOptionAdmin(c echo.Context) error {
	var req createOptionAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	option := models.FieldOption{
		Name:      req.Name,
		FieldID:   req.FieldID,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	}
	item, err := s.store.CreateOption(c.Request().Context(), option)
	if errors.Is(err, storage.ErrFieldNotFound) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": storage.ErrFieldNotFound.Error()})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "create option error"})
	}
	return c.JSON(http.StatusCreated, makeOptionResponse(item))
}

func (s *Server) deleteOptionAdmin(c echo.Context) error {
	optionID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	err = s.store.DeleteOption(c.Request().Context(), optionID)
	if errors.Is(err, storage.ErrOptionNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "option not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "delete option error"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) updateOptionAdmin(c echo.Context) error {
	optionID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	var req createFieldOptionAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	option := models.FieldOption{
		ID:        optionID,
		Name:      req.Name,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	}
	item, err := s.store.UpdateOption(c.Request().Context(), option)
	if errors.Is(err, storage.ErrOptionNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "option not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update option error"})
	}
	return c.JSON(http.StatusOK, makeOptionResponse(item))
}
