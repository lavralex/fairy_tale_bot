package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lavralex/fairy_tale_bot/internal/config"
)

type Server struct {
	echo  *echo.Echo
	conf  *config.Config
	store Storage
}

func New(conf *config.Config, store Storage) *Server {
	return &Server{
		echo:  echo.New(),
		conf:  conf,
		store: store,
	}
}

func (s *Server) Run(ctx context.Context) error {
	e := s.echo
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	s.setupRoutes()

	errCh := make(chan error, 1)

	go func() { errCh <- e.Start(":8080") }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := e.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("server shutdown error: %v", err)
		}
		cancel()
		return ctx.Err()
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("server error: %w", err)

	}
}

func (s *Server) setupRoutes() {
	s.echo.GET("/", hello)
	s.echo.POST("/admin/products", s.createProductAdmin)
	s.echo.GET("/admin/products/:id", s.getProductAdmin)
	s.echo.DELETE("/admin/products/:id", s.deleteProductAdmin)
	s.echo.GET("/admin/products", s.listProducts)
	s.echo.PUT("/admin/products/:id", s.updateProductAdmin)
	s.echo.POST("/admin/fields", s.createFieldAdmin)
	s.echo.GET("/admin/fields/:id", s.getFieldAdmin)
	s.echo.GET("/admin/fields", s.listFieldsAdmin)
	s.echo.PUT("/admin/fields/:id", s.updateFieldAdmin)
	s.echo.DELETE("/admin/fields/:id", s.deleteFieldAdmin)
	s.echo.POST("/admin/options", s.createOptionAdmin)
	s.echo.PUT("/admin/options/:id", s.updateOptionAdmin)
	s.echo.DELETE("/admin/options/:id", s.deleteOptionAdmin)
	s.echo.POST("/admin/templates", s.createTemplateAdmin)
	s.echo.GET("/admin/templates/:id", s.getTemplateAdmin)
	s.echo.GET("/admin/templates", s.listTemplatesAdmin)
	s.echo.DELETE("/admin/templates/:id", s.deleteTemplateAdmin)
	s.echo.PUT("/admin/templates/:id", s.updateTemplateAdmin)
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
