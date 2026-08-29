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

type createTemplateAdminRequest struct {
	Name   string                            `json:"name"`
	Fields []createTemplateFieldAdminRequest `json:"fields"`
}
type createTemplateFieldAdminRequest struct {
	ID int64 `json:"id"`
}

type templateAdminResponse struct {
	ID        int64                        `json:"id"`
	Name      string                       `json:"name"`
	CreatedAt time.Time                    `json:"created_at"`
	Fields    []templateFieldAdminResponse `json:"fields"`
}

type templateFieldAdminResponse struct {
	Field     fieldAdminResponse `json:"field"`
	SortOrder int                `json:"sort_order"`
}

func makeTemplateResponse(template models.Template) templateAdminResponse {
	templateFieldsResponses := make([]templateFieldAdminResponse, 0, len(template.Fields))
	for _, tf := range template.Fields {
		templateFieldsResponses = append(
			templateFieldsResponses,
			templateFieldAdminResponse{
				Field:     makeFieldResponse(tf.Field),
				SortOrder: tf.SortOrder,
			},
		)
	}
	return templateAdminResponse{
		ID:        template.ID,
		Name:      template.Name,
		CreatedAt: template.CreatedAt,
		Fields:    templateFieldsResponses,
	}
}

func (s *Server) createTemplateAdmin(c echo.Context) error {
	var req createTemplateAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	template := models.Template{
		Name: req.Name,
	}
	for order, field := range req.Fields {

		template.Fields = append(
			template.Fields,
			models.TemplateField{
				Field: models.Field{
					ID: field.ID,
				},
				SortOrder: order,
			},
		)
	}
	item, err := s.store.CreateTemplate(c.Request().Context(), template)
	if errors.Is(err, storage.ErrIncorrectNumberOfFields) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "the incorrect number of fields: required is 1 and more",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "create template error"})
	}
	return c.JSON(http.StatusCreated, makeTemplateResponse(item))
}

func (s *Server) getTemplateAdmin(c echo.Context) error {
	templateID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	item, err := s.store.GetTemplate(c.Request().Context(), templateID)
	if errors.Is(err, storage.ErrTemplateNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "template not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "get template error"})
	}
	return c.JSON(http.StatusOK, makeTemplateResponse(item))
}

type listTemplateResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

const templatesLimit = 50

func (s *Server) listTemplatesAdmin(c echo.Context) error {
	limit := parseLimit(c, templatesLimit)
	offset := parseOffset(c, limit)
	items, err := s.store.GetTemplates(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "get templates error"})
	}
	templates := make([]listTemplateResponse, 0, len(items))
	for _, item := range items {
		templates = append(
			templates,
			listTemplateResponse{
				ID:        item.ID,
				Name:      item.Name,
				CreatedAt: item.CreatedAt,
			},
		)
	}
	return c.JSON(http.StatusOK, templates)
}

func (s *Server) deleteTemplateAdmin(c echo.Context) error {
	templateID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	err = s.store.DeleteTemplate(c.Request().Context(), templateID)
	if errors.Is(err, storage.ErrTemplateNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "template not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "delete template error"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) updateTemplateAdmin(c echo.Context) error {
	templateID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	var req createTemplateAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	template := models.Template{
		ID:   templateID,
		Name: req.Name,
	}
	for order, field := range req.Fields {
		template.Fields = append(
			template.Fields,
			models.TemplateField{
				Field: models.Field{
					ID: field.ID,
				},
				SortOrder: order,
			},
		)
	}
	item, err := s.store.UpdateTemplate(c.Request().Context(), template)
	if errors.Is(err, storage.ErrTemplateNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "template not found"})
	}
	if errors.Is(err, storage.ErrIncorrectNumberOfFields) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "the incorrect number of fields: required is 1 and more",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update template error"})
	}
	return c.JSON(http.StatusOK, makeTemplateResponse(item))
}
