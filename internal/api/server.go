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
)

type Server struct {
	echo *echo.Echo
	conf *config.Config
}

func New(conf *config.Config) *Server {
	return &Server{
		echo: echo.New(),
		conf: conf,
	}
}

func (s *Server) Run(ctx context.Context) error {
	e := s.echo
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.GET("/", hello)

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

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
