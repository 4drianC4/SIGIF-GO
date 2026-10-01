package dtos

import (
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UserResponse struct {
	ID        string                `json:"id"`
	TenantID  string                `json:"tenant_id"`
	Email     string                `json:"email"`
	FirstName string                `json:"first_name"`
	LastName  string                `json:"last_name"`
	FullName  string                `json:"full_name"`
	Roles     []userEntity.UserRole `json:"roles"`
	Status    userEntity.UserStatus `json:"status"`
}

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type AuthResponse struct {
	User   UserResponse      `json:"user"`
	Tokens TokenPairResponse `json:"tokens"`
}

func FromUser(u *userEntity.User, tokens interface {
	AccessToken() string
	RefreshToken() string
	ExpiresIn() int
	TokenType() string
}) AuthResponse {
	return AuthResponse{
		User: UserResponse{
			ID:        u.ID.String(),
			TenantID:  u.TenantID.String(),
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			FullName:  u.FullName(),
			Roles:     u.Roles,
			Status:    u.Status,
		},
		Tokens: TokenPairResponse{
			AccessToken:  tokens.AccessToken(),
			RefreshToken: tokens.RefreshToken(),
			ExpiresIn:    tokens.ExpiresIn(),
			TokenType:    tokens.TokenType(),
		},
	}
}
