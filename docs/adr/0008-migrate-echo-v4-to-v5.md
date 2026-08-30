# ADR 0008: Framework Upgrade — Migrating REST API Service to Echo v5

## Status

Accepted

## Context

The REST API microservice in `backend` was built on `github.com/labstack/echo/v4`. While Echo v4 is supported in maintenance mode until December 31, 2026, Echo v5 introduces several major framework improvements:
1. **Concrete Context Struct:** Transitioning `echo.Context` from an interface to a concrete struct pointer (`*echo.Context`) enables adding methods without breaking interface compatibility in future minor versions.
2. **Standard Library `log/slog` Integration:** Native adoption of `log/slog` aligns Echo with modern Go 1.21+ logging conventions and eliminates legacy logging dependencies.
3. **Improved Router & Lifecycle Management:** Replacing legacy `e.Start()`/`e.Shutdown()` methods with unified `echo.StartConfig` simplifies context-aware graceful shutdown and server configuration.
4. **Enhanced Type Safety & Helpers:** Updated parameter helpers (`echo.PathParam`, `c.SetPathValues`) and error wrapping APIs improve API developer ergonomics.

Upgrading to Echo v5 ensures long-term framework support, improved performance, and alignment with modern Go standard library patterns.

## Decision

We will upgrade the `backend` microservices suite from `github.com/labstack/echo/v4` to `github.com/labstack/echo/v5` and `samber/slog-echo/v2`.

### Migration Standards & Key Changes

1. **Context Struct Pointer (`*echo.Context`):**
   All HTTP handler functions and custom middleware signatures MUST accept `*echo.Context` instead of `echo.Context`.
   ```go
   // Handlers signature
   func MyHandler(svc *service.MyService) echo.HandlerFunc {
       return func(c *echo.Context) error { ... }
   }

   // Middleware signature
   func MyMiddleware() echo.MiddlewareFunc {
       return func(next echo.HandlerFunc) echo.HandlerFunc {
           return func(c *echo.Context) error { ... }
       }
   }
   ```

2. **HTTP Error Handler Parameter Order:**
   The `echo.HTTPErrorHandler` function signature has been reversed to `func(c *echo.Context, err error)`. `pkg/apperror/echo.go` has been refactored accordingly:
   ```go
   func NewEchoHTTPErrorHandler() echo.HTTPErrorHandler {
       return func(c *echo.Context, err error) { ... }
   }
   ```

3. **HTTPError Message Field:**
   In Echo v5, `echo.HTTPError.Message` is typed directly as a `string` (previously `interface{}`). Type assertions on `he.Message` are no longer required.

4. **Server Lifecycle via `echo.StartConfig`:**
   In `cmd/api.go`, server startup and context-cancellation graceful shutdown are now managed via `echo.StartConfig`:
   ```go
   sc := echo.StartConfig{
       Address:         fmt.Sprintf(":%d", cfg.Server.Port),
       GracefulTimeout: 10 * time.Second,
   }
   if err := sc.Start(ctx, e); err != nil && !errors.Is(err, http.ErrServerClosed) {
       slog.Error("server startup failed", "error", err)
       return err
   }
   ```

5. **Testing Path Parameter Injection (`c.SetPathValues`):**
   In unit tests, path parameters on `*echo.Context` are set using `c.SetPathValues(echo.PathValues{echo.PathValue{Name: "app_id", Value: appID.String()}})`.

## Consequences

### Positive

- **Future Compatibility:** Positions the codebase on the actively developed Echo v5 major release line.
- **Native Structured Logging:** Direct integration with `log/slog` across Echo's internal logging and HTTP middleware.
- **Simplified Lifecycle:** Declarative graceful shutdown via `echo.StartConfig` eliminates manual goroutine management and timeout contexts in CLI commands.
- **Type Safety & Ergonomics:** Concrete `*echo.Context` struct avoids interface method overhead and enables future non-breaking method additions.

### Negative

- **Breaking API Changes:** Required updating every HTTP handler, middleware, error handler, and router setup file across `internal/api/` and `pkg/apperror/`.
