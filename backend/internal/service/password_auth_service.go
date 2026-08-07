package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type customClaims struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	UserID   string
	jwt.RegisteredClaims
}

const (
	SubjectPasswordAuth  = "static:admin"
	AdminEmailClaimValue = "admin@admin.com"
)

type PasswordAuthRepository interface {
	ProvisionUser(ctx context.Context, userData *model.UserProvisionData) (*public.User, error)
	GetAdminApplications(ctx context.Context) ([]public.Application, error)
}

type PasswordAuthService struct {
	authConfig *config.AuthConfig
	repo       PasswordAuthRepository
}

func NewPasswordAuthService(authConfig *config.AuthConfig, repo PasswordAuthRepository) *PasswordAuthService {
	return &PasswordAuthService{
		authConfig: authConfig,
		repo:       repo,
	}
}

func (p *PasswordAuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	if subtle.ConstantTimeCompare([]byte(p.authConfig.Admin.Username), []byte(username)) != 1 ||
		subtle.ConstantTimeCompare([]byte(p.authConfig.Admin.Password), []byte(password)) != 1 {
		return "", fmt.Errorf("username or password do not match")
	}

	u, err := p.repo.ProvisionUser(ctx, &model.UserProvisionData{
		Email:         AdminEmailClaimValue,
		EmailVerified: true,
		Subject:       SubjectPasswordAuth,
	})

	c := customClaims{
		Username: p.authConfig.Admin.Username,
		Email:    AdminEmailClaimValue,
		UserID:   u.ID.String(),
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

	if err != nil {
		return "", err
	}

	return jwtToken, err
}

func (p *PasswordAuthService) VerifyToken(ctx context.Context, token string) (*model.UserProfile, error) {
	claims := &customClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.authConfig.Admin.JWTSecret), nil
	})

	if err != nil {
		// Handle specific validation errors (e.g., expired or invalid signatures)
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("token signature is invalid")
		}
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	// Final sanity check to ensure the token object is fully marked valid
	if !t.Valid {
		return nil, errors.New("invalid token")
	}
	return &model.UserProfile{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Subject:  claims.Subject,
	}, nil
}

func (p *PasswordAuthService) PermittedApplications(ctx context.Context, userID string) ([]public.Application, error) {
	if userID == "" {
		return nil, fmt.Errorf("invliad userID")
	}
	_, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	apps, err := p.repo.GetAdminApplications(ctx)
	if err != nil {
		return nil, err
	}
	return apps, nil
}
