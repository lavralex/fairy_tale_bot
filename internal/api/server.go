package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lavralex/fairy_tale_bot/internal/config"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

type Server struct {
	echo  *echo.Echo
	conf  *config.Config
	store *storage.Storage
}

func New(conf *config.Config, store *storage.Storage) *Server {
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
		e.Shutdown(shutdownCtx)
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
	s.echo.GET("/admin/products", s.listProducts)
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
