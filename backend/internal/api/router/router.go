package router

import (
	"log/slog"

	"github.com/ajaxe/email-ingestion/internal/api/handler"
	"github.com/ajaxe/email-ingestion/internal/api/middleware"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

func New(cfg *config.AppConfig) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Generic middleware
	e.Use(slogecho.New(slog.Default()))
	e.Use(echomiddleware.Recover())

	// Edge API group
	edgeGroup := e.Group("/internal/api/v1")
	edgeGroup.Use(middleware.EdgeAuth(cfg.Smtp.MTAAuthToken))
	edgeGroup.POST("/ingest", handler.HandleIngest)

	return e
}
