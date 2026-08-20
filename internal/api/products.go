package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lavralex/fairy_tale_bot/internal/models"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

type createProductAdminRequest struct {
	Name            string                    `json:"name"`
	Description     string                    `json:"description"`
	Price           string                    `json:"price"`
	Tags            []string                  `json:"tags,omitempty"`
	IsAvailable     bool                      `json:"is_available"`
	CommentEnabled  bool                      `json:"comment_enabled"`
	TemplateID      *int64                    `json:"template_id,omitempty"`
	TemplateEnabled bool                      `json:"template_enabled"`
	Images          []createImageAdminRequest `json:"images"`
}

type createImageAdminRequest struct {
	Src       string `json:"src"`
	Alt       string `json:"alt"`
	SortOrder int8   `json:"sort_order"`
}

type productAdminResponse struct {
	ID              int64                `json:"id"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Price           string               `json:"price"`
	Tags            []string             `json:"tags,omitempty"`
	TemplateID      *int64               `json:"template_id,omitempty"`
	IsAvailable     bool                 `json:"is_available"`
	CommentEnabled  bool                 `json:"comment_enabled"`
	TemplateEnabled bool                 `json:"template_enabled"`
	CreatedAt       time.Time            `json:"created_at"`
	Images          []imageAdminResponse `json:"images"`
}

type imageAdminResponse struct {
	ID        int64     `json:"id"`
	Src       string    `json:"src"`
	Alt       string    `json:"alt"`
	SortOrder int8      `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) createProductAdmin(c echo.Context) error {
	var req createProductAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	product := models.Product{
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		Tags:            req.Tags,
		IsAvailable:     req.IsAvailable,
		CommentEnabled:  req.CommentEnabled,
		TemplateID:      req.TemplateID,
		TemplateEnabled: req.TemplateEnabled,
	}
	for _, image := range req.Images {
		product.Images = append(
			product.Images,
			models.ProductImage{
				Src:       image.Src,
				Alt:       image.Alt,
				SortOrder: image.SortOrder,
			},
		)
	}
	item, err := s.store.CreateProduct(c.Request().Context(), product)
	if errors.Is(err, storage.ErrIncorrectNumberOfImages) {
		errTxt := fmt.Sprintf("the incorrect number of images: %d required is between 1 and 5", len(product.Images))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errTxt})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "create product error"})
	}
	createdImages := make([]imageAdminResponse, 0, len(item.Images))
	for _, image := range item.Images {
		createdImages = append(
			createdImages,
			imageAdminResponse{
				ID:        image.ID,
				Src:       image.Src,
				Alt:       image.Alt,
				SortOrder: image.SortOrder,
				CreatedAt: image.CreatedAt,
			},
		)
	}
	createdProduct := productAdminResponse{
		ID:              item.ID,
		Name:            item.Name,
		Description:     item.Description,
		Price:           item.Price,
		Tags:            item.Tags,
		IsAvailable:     item.IsAvailable,
		CommentEnabled:  item.CommentEnabled,
		TemplateID:      item.TemplateID,
		TemplateEnabled: item.TemplateEnabled,
		CreatedAt:       item.CreatedAt,
		Images:          createdImages,
	}
	return c.JSON(http.StatusCreated, createdProduct)
}

func (s *Server) getProductAdmin(c echo.Context) error {
	productID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	item, err := s.store.GetProduct(c.Request().Context(), productID)
	if errors.Is(err, storage.ErrProductNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "product not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "get product error"})
	}
	images := make([]imageAdminResponse, 0, len(item.Images))
	for _, image := range item.Images {
		images = append(
			images,
			imageAdminResponse{
				ID:        image.ID,
				Src:       image.Src,
				Alt:       image.Alt,
				SortOrder: image.SortOrder,
				CreatedAt: image.CreatedAt,
			},
		)
	}
	product := productAdminResponse{
		ID:              item.ID,
		Name:            item.Name,
		Description:     item.Description,
		Price:           item.Price,
		TemplateID:      item.TemplateID,
		Tags:            item.Tags,
		CreatedAt:       item.CreatedAt,
		IsAvailable:     item.IsAvailable,
		CommentEnabled:  item.CommentEnabled,
		TemplateEnabled: item.TemplateEnabled,
		Images:          images,
	}
	return c.JSON(http.StatusOK, product)
}

type listProductsResponse struct {
	ID          int64                      `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Price       string                     `json:"price"`
	Tags        []string                   `json:"tags,omitempty"`
	Images      []listProductImageResponse `json:"images"`
}

type listProductImageResponse struct {
	ID        int64  `json:"id"`
	Src       string `json:"src"`
	Alt       string `json:"alt"`
	SortOrder int8   `json:"sort_order"`
}

const defaultLimit = 10
const maxLimit = 50

func parseLimit(c echo.Context) int {
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

const defaultOffset = 0

func parseOffset(c echo.Context, limit int) int {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page <= 0 {
		return defaultOffset
	}
	return (page - 1) * limit
}

func (s *Server) listProducts(c echo.Context) error {
	limit := parseLimit(c)
	offset := parseOffset(c, limit)
	items, err := s.store.GetProducts(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "get products error"})
	}
	products := make([]listProductsResponse, 0, limit)
	for _, item := range items {
		respImages := make([]listProductImageResponse, 0, len(item.Images))
		for _, image := range item.Images {
			respImages = append(
				respImages,
				listProductImageResponse{
					ID:        image.ID,
					Src:       image.Src,
					Alt:       image.Alt,
					SortOrder: image.SortOrder,
				},
			)
		}
		products = append(
			products,
			listProductsResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Price:       item.Price,
				Tags:        item.Tags,
				Images:      respImages,
			},
		)
	}
	return c.JSON(http.StatusOK, products)
}

func (s *Server) deleteProductAdmin(c echo.Context) error {
	productID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	err = s.store.DeleteProduct(c.Request().Context(), productID)
	if errors.Is(err, storage.ErrProductNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "product not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "delete product error"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) updateProductAdmin(c echo.Context) error {
	productID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id parameter can't parse to number error"})
	}
	var req createProductAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	product := models.Product{
		ID:              productID,
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		Tags:            req.Tags,
		IsAvailable:     req.IsAvailable,
		CommentEnabled:  req.CommentEnabled,
		TemplateID:      req.TemplateID,
		TemplateEnabled: req.TemplateEnabled,
	}
	for _, image := range req.Images {
		product.Images = append(
			product.Images,
			models.ProductImage{
				Src:       image.Src,
				Alt:       image.Alt,
				SortOrder: image.SortOrder,
			},
		)
	}
	item, err := s.store.UpdateProduct(c.Request().Context(), product)
	if errors.Is(err, storage.ErrProductNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "product not found"})
	}
	if errors.Is(err, storage.ErrIncorrectNumberOfImages) {
		errTxt := fmt.Sprintf("the incorrect number of images: %d required is between 1 and 5", len(product.Images))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errTxt})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update product error"})
	}
	updatedImages := make([]imageAdminResponse, 0, len(item.Images))
	for _, image := range item.Images {
		updatedImages = append(
			updatedImages,
			imageAdminResponse{
				ID:        image.ID,
				Src:       image.Src,
				Alt:       image.Alt,
				SortOrder: image.SortOrder,
				CreatedAt: image.CreatedAt,
			},
		)
	}
	updatedProduct := productAdminResponse{
		ID:              item.ID,
		Name:            item.Name,
		Description:     item.Description,
		Price:           item.Price,
		Tags:            item.Tags,
		IsAvailable:     item.IsAvailable,
		CommentEnabled:  item.CommentEnabled,
		TemplateID:      item.TemplateID,
		TemplateEnabled: item.TemplateEnabled,
		CreatedAt:       item.CreatedAt,
		Images:          updatedImages,
	}
	return c.JSON(http.StatusOK, updatedProduct)
}
