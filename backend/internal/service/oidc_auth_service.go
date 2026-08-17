package service

import (
	"context"
	"fmt"

	"github.com/ajaxe/email-ingestion/internal/api/dto"
	"github.com/ajaxe/email-ingestion/pkg/apperror"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
)

type OIDCAuthRepository interface {
	ProvisionUser(ctx context.Context, userData *UserProvisionData) (*public.User, error)
	GetApplications(ctx context.Context, userID uuid.UUID) ([]public.Application, error)
}

type oidcUserProfileClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

type OIDCAuthService struct {
	authConfig *config.AuthConfig
	repo       OIDCAuthRepository
	verifier   *oidc.IDTokenVerifier
}

func NewOIDCAuthService(authConfig *config.AuthConfig, repo OIDCAuthRepository) *OIDCAuthService {
	keyset := oidc.NewRemoteKeySet(context.Background(), authConfig.OIDC.JWKSURI)
	oidcVerifier := oidc.NewVerifier(authConfig.OIDC.Authority, keyset, &oidc.Config{
		ClientID: authConfig.OIDC.ClientID,
	})
	return &OIDCAuthService{
		verifier:   oidcVerifier,
		authConfig: authConfig,
		repo:       repo,
	}
}

func (o *OIDCAuthService) VerifyToken(ctx context.Context, token string) (up *dto.UserProfile, err error) {
	idToken, err := o.verifier.Verify(ctx, token)
	if err != nil {
		err = apperror.Forbidden("forbidden_token_invalid", err)
		return
	}
	var profile oidcUserProfileClaims
	if err = idToken.Claims(&profile); err != nil {
		err = apperror.Forbidden("forbidden_token_invalid", err)
		return
	}

	provisionData := &UserProvisionData{
		Email:   profile.Email,
		Subject: profile.Sub,
	}

	// lazy provision of partial user info in the database
	user, err := o.repo.ProvisionUser(ctx, provisionData)
	if err != nil {
		err = apperror.Forbidden("forbidden_user", err)
		return
	}

	up = &dto.UserProfile{
		Username: user.Email,
		Email:    user.Email,
		Subject:  user.IdpUserSub,
		UserID:   user.ID.String(),
	}
	return
}

func (o *OIDCAuthService) PermittedApplications(ctx context.Context, userID string) (apps []public.Application, err error) {
	defer func() {
		if err != nil {
			apps = nil
			e := fmt.Errorf("failed to get permitted applications: %w", err)
			err = apperror.Forbidden("forbidden", e)
		}
	}()
	if userID == "" {
		err = fmt.Errorf("invalid userID")
		return
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		err = fmt.Errorf("invalid userID: %w", err)
		return
	}
	if uid == uuid.Nil {
		err = fmt.Errorf("invalid user id, empty uuid")
		return
	}

	apps, err = o.repo.GetApplications(ctx, uid)
	if err != nil {
		return
	}
	if apps == nil {
		apps = []public.Application{}
	}
	return
}
