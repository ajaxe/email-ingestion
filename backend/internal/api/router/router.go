package router

import (
	"log/slog"

	"github.com/ajaxe/email-ingestion/internal/api/handler"
	"github.com/ajaxe/email-ingestion/internal/api/middleware"
	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

type ApiInitOptions struct {
	// Add any initialization options here
	Queries      *public.Queries
	RedisManager *redis.Manager
	DBPool       *pgxpool.Pool
}

func New(cfg *config.AppConfig, o *ApiInitOptions) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = apperror.NewEchoHTTPErrorHandler()

	// Generic middleware
	e.Use(slogecho.New(slog.Default()))
	e.Use(echomiddleware.Recover())

	storageService := storage.NewStorageService(&cfg.Storage)
	apiKeyService := service.NewApiKeyService(o.Queries)

	configureInternalAPI(e, cfg, o, storageService)

	configureAppAPI(e, cfg, o, apiKeyService, storageService)

	configureM2MAPI(e, o, storageService, apiKeyService)

	return e
}

func configureM2MAPI(e *echo.Echo, o *ApiInitOptions, storageService *storage.S3StorageService, apiKeyService *service.ApiKeyService) {
	authz := service.NewAuthorizationService(o.Queries, o.RedisManager.Cache, apiKeyService)
	// M2M API routes
	m2mGroup := e.Group("/api/v1")
	m2mGroup.Use(middleware.M2MAuth(authz))

	emailService := service.NewEmailService(o.Queries, storageService)

	m2mGroup.GET("/emails/:email_id", handler.HandleAPIEmailByID(emailService))
	m2mGroup.GET("/emails/:email_id/attachments/:attachment_id", handler.HandleAPIGetAttachmentURL(emailService))
}

func configureAppAPI(e *echo.Echo, cfg *config.AppConfig, o *ApiInitOptions, apiKeyService *service.ApiKeyService, storageService *storage.S3StorageService) {
	// Application API
	// TODO: When Phase 6.1 is implemented, this should move to a JWT-protected /api/v1 group.
	// For now, mapping it here for testing Phase 5.1
	webhookStreamName := util.WebhookStreamName(cfg.Environment)
	var publisher service.EventPublisher
	if o.RedisManager != nil {
		publisher = o.RedisManager.Stream
	}
	webhookService := service.NewWebhookService(o.Queries, &cfg.Webhook, publisher, webhookStreamName)
	authz := service.NewAuthorizationService(o.Queries, o.RedisManager.Cache, apiKeyService)
	pwdAuthService := service.NewPasswordAuthService(&cfg.Auth, service.NewAppPasswordAuthRepository(o.Queries, authz))

	var authService middleware.UserAccessVerifier

	if cfg.Auth.Provider == config.PasswordAuthProvider {
		authService = pwdAuthService
	} else {
		authService = service.NewOIDCAuthService(&cfg.Auth, service.NewPgxOIDCAuthRepository(o.Queries, authz))
	}

	prefix := "/app/v1"
	appGroup := e.Group(prefix)
	appGroup.Use(middleware.AppAuth(authService))

	appService := service.NewApplicationService(service.NewPgxApplicationRepository(o.DBPool))
	emailService := service.NewEmailService(o.Queries, storageService)
	appGroup.GET("/applications", handler.HandleGetApplications(appService))
	appGroup.POST("/applications", handler.HandleCreateApplication(appService))
	appGroup.GET("/applications/:app_id", handler.HandleGetApplicationByID(appService))

	appGroup.GET("/applications/:app_id/stats", handler.HandleGetApplicationStats(appService))

	appGroup.GET("/applications/:app_id/addresses", handler.HandleListAddresses(appService))
	appGroup.POST("/applications/:app_id/addresses", handler.HandleCreateAddress(appService))
	appGroup.PATCH("/applications/:app_id/addresses/:address_id", handler.HandleToggleAddressStatus(appService))

	appGroup.GET("/applications/:app_id/emails", handler.HandleListEmails(appService))
	appGroup.GET("/applications/:app_id/emails/:email_id", handler.HandleGetEmailByID(emailService))
	appGroup.GET("/applications/:app_id/emails/:email_id/attachments/:attachment_id", handler.HandleGetAttachmentURL(emailService))

	appGroup.GET("/applications/:app_id/api-keys", handler.HandleListAPIKeys(apiKeyService))
	appGroup.POST("/applications/:app_id/api-keys", handler.HandleCreateAPIKey(apiKeyService))
	appGroup.DELETE("/applications/:app_id/api-keys/:key_id", handler.HandleRevokeAPIKey(apiKeyService))

	appGroup.POST("/applications/:app_id/webhook", handler.HandleRegisterWebhook(webhookService))
	appGroup.PUT("/applications/:app_id/webhook", handler.HandlePutUpdateWebhook(webhookService))
	appGroup.GET("/applications/:app_id/webhook/jobs", handler.HandleListWebhookJobs(webhookService))
	appGroup.POST("/applications/:app_id/webhook/jobs/:job_id/redeliver", handler.HandleRedeliverWebhookJob(webhookService))

	appGroup.GET("/auth/session", handler.HandleGetAuthSession())

	// TODO: open endpoint needs protection, may be move to SPA as static file, need to address maintenance of the file.
	e.GET(prefix+"/auth/config", handler.HandleGetAuthConfig(&cfg.Auth, cfg.Smtp.EmailDomain))

	// TODO: login endpoint is strictly for internal usage, there are no protections on this open endpoint.
	e.POST(prefix+"/auth/login", handler.HandlePostLogin(pwdAuthService))

}

// configureInternalAPI sets up the internal API routes for the application, including the ingestion and email lookup endpoints.
// begins with api prefix: /internal/api/v1
func configureInternalAPI(e *echo.Echo, cfg *config.AppConfig, o *ApiInitOptions, storageService *storage.S3StorageService) {
	ingestionService := service.NewEmailIngestion(o.RedisManager, o.Queries, storageService, cfg.Environment)

	// Edge API group
	edgeGroup := e.Group("/internal/api/v1")
	edgeGroup.Use(middleware.EdgeAuth(cfg.Smtp.MTAAuthToken))
	edgeGroup.POST("/ingest", handler.HandleIngest(ingestionService))
	edgeGroup.GET("/addresses/:email", handler.HandleIngestEmailLookup("email", ingestionService))
}
