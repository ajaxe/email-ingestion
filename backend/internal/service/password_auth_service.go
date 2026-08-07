package service

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

type customClaims struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

const (
	SubjectPasswordAuth  = "static:admin"
	AdminEmailClaimValue = "admin@admin.com"
)

type PasswordAuthService struct {
	authConfig *config.AuthConfig
	authz      *AuthorizationService
}

func NewPasswordAuthService(authConfig *config.AuthConfig, authz *AuthorizationService) *PasswordAuthService {
	return &PasswordAuthService{
		authConfig: authConfig,
		authz:      authz,
	}
}

func (p *PasswordAuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	if subtle.ConstantTimeCompare([]byte(p.authConfig.Admin.Username), []byte(username)) != 1 ||
		subtle.ConstantTimeCompare([]byte(p.authConfig.Admin.Password), []byte(password)) != 1 {
		return "", fmt.Errorf("username or password do not match")
	}
	c := customClaims{
		Username: p.authConfig.Admin.Username,
		Email:    AdminEmailClaimValue,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(p.authConfig.Admin.TokenTTLMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    p.authConfig.Admin.Issuer,
			Subject:   SubjectPasswordAuth,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	jwtToken, err := token.SignedString([]byte(p.authConfig.Admin.JWTSecret))

	if err != nil {
		return "", err
	}

	err = p.authz.ProvisionUser(ctx, &model.UserProvisionData{
		Email:         AdminEmailClaimValue,
		EmailVerified: true,
		Subject:       SubjectPasswordAuth,
	})

	if err != nil {
		return "", err
	}

	return jwtToken, err
}
