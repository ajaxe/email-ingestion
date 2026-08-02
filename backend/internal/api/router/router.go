package router

import (
	"log/slog"

	"github.com/ajaxe/email-ingestion/internal/api/handler"
	"github.com/ajaxe/email-ingestion/internal/api/middleware"
	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

type ApiInitOptions struct {
	// Add any initialization options here
	Queries      *public.Queries
	RedisManager *redis.Manager
}

func New(cfg *config.AppConfig, o *ApiInitOptions) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Generic middleware
	e.Use(slogecho.New(slog.Default()))
	e.Use(echomiddleware.Recover())

	storageService := storage.NewStorageService(&cfg.Storage)

	emailService := service.NewEmailIngestion(o.RedisManager, o.Queries, storageService, cfg.Environment)

	// Edge API group
	edgeGroup := e.Group("/internal/api/v1")
	edgeGroup.Use(middleware.EdgeAuth(cfg.Smtp.MTAAuthToken))
	edgeGroup.POST("/ingest", handler.HandleIngest(emailService))
	edgeGroup.GET("/addresses/:email", handler.HandleIngestEmailLookup("email", emailService))

	return e
}
