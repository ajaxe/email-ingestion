package service

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

type customClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

const (
	SubjectPasswordAuth = "static:admin"
)

type PasswordAuthService struct {
	authConfig *config.AuthConfig
}

func NewPasswordAuthService(authConfig *config.AuthConfig) *PasswordAuthService {
	return &PasswordAuthService{
		authConfig: authConfig,
	}
}

func (p *PasswordAuthService) Authenticate(username, password string) (string, error) {
	if subtle.ConstantTimeCompare([]byte(p.authConfig.Admin.Username), []byte(username)) != 1 ||
		subtle.ConstantTimeCompare([]byte(p.authConfig.Admin.Password), []byte(password)) != 1 {
		return "", fmt.Errorf("username or password do not match")
	}
	c := customClaims{
		Username: p.authConfig.Admin.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(p.authConfig.Admin.TokenTTLMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    p.authConfig.Admin.Issuer,
		},
	}
	c.Subject = SubjectPasswordAuth
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	jwtToken, err := token.SignedString([]byte(p.authConfig.Admin.JWTSecret))

	if err != nil {
		return "", err
	}

	return jwtToken, err
}
