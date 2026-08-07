package handler

import (
	"encoding/json"
	"net/http"
	"text/template"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/labstack/echo/v4"
)

const authConfigTemplate = `window.APP_CONFIG = {
  AUTH_PROVIDER: '{{ .Provider }}',{{if eq .Provider "oidc"}}
  OIDC_AUTHORITY: '{{ .OIDC.Authority }}',
  OIDC_CLIENT_ID: '{{ .OIDC.ClientID }}',{{end}}
}`

var tmpl = template.Must(template.New("auth_config").Parse(authConfigTemplate))

func HandleGetAuthConfig(authCfg *config.AuthConfig) echo.HandlerFunc {
	return func(e echo.Context) error {
		e.Response().Header().Set("Cache-Control", "public, max-age=3600")
		e.Response().Header().Set(echo.HeaderContentType, "application/javascript; charset=utf-8")
		e.Response().WriteHeader(http.StatusOK)
		return tmpl.Execute(e.Response(), authCfg)
	}
}

func HandlePostLogin(authService *service.PasswordAuthService) echo.HandlerFunc {
	return func(e echo.Context) error {
		req := &dto.LoginAuthRequest{}
		if err := json.NewDecoder(e.Request().Body).Decode(req); err != nil {
			return apperror.Validation("Invalid login request", err)
		}
		token, err := authService.Authenticate(e.Request().Context(), req.Username, req.Password)
		if err != nil {
			return apperror.Unauthorized("Invalid username & password", err)
		}
		return e.JSON(http.StatusOK, map[string]string{
			"token": token,
		})
	}
}
