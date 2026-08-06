package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/labstack/echo/v4"
)

func HandleGetAuthConfig(authCfg *config.AuthConfig) echo.HandlerFunc {
	return func(e echo.Context) error {
		e.Response().Header().Set("Cache-Control", "public, max-age=3600")
		if authCfg.Provider == "password" {
			return e.JSON(200, map[string]string{
				"provider": authCfg.Provider,
			})
		}
		return e.JSON(200, map[string]string{
			"provider":  authCfg.Provider,
			"client_id": authCfg.OIDC.ClientID,
			"issuer":    authCfg.OIDC.Issuer,
		})
	}
}

func HandlePostLogin(authService *service.PasswordAuthService) echo.HandlerFunc {
	return func(e echo.Context) error {
		req := &model.LoginAuthRequest{}
		if err := json.NewDecoder(e.Request().Body).Decode(req); err != nil {
			return apperror.Validation("Invalid login request", err)
		}
		token, err := authService.Authenticate(req.Username, req.Password)
		if err != nil {
			return apperror.Unauthorized("Invalid username & password", err)
		}
		return e.JSON(http.StatusOK, map[string]string{
			"token": token,
		})
	}
}
