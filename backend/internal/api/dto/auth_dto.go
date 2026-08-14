package dto

import (
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/google/uuid"
)

type LoginAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserProfile struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Subject  string `json:"subject"`
	Email    string `json:"email"`
}

type UserAccessResult struct {
	UserProfile  *UserProfile         `json:"userProfile"`
	Applications []public.Application `json:"applications"`
}

func (u *UserAccessResult) ApplicationByID(id uuid.UUID) *public.Application {
	for _, app := range u.Applications {
		if app.ID == id {
			return &app
		}
	}
	return nil
}

func (u *UserAccessResult) CanAccessApplication(id uuid.UUID) bool {
	return u.ApplicationByID(id) != nil
}
