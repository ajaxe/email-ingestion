package router

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/ajaxe/email-ingestion/internal/api/handler"
	"github.com/ajaxe/email-ingestion/internal/api/middleware"
)

func New(edgeToken string) *echo.Echo {
	e := echo.New()

	// Generic middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	// Edge API group
	edgeGroup := e.Group("/internal/api/v1")
	edgeGroup.Use(middleware.EdgeAuth(edgeToken))
	edgeGroup.POST("/ingest", handler.HandleIngest)

	return e
}
