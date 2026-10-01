package dto

import (
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/jwt"
)

type User struct {
	ID        string                `json:"id"`
	TenantID  string                `json:"tenant_id"`
	Email     string                `json:"email"`
	FirstName string                `json:"first_name"`
	LastName  string                `json:"last_name"`
	FullName  string                `json:"full_name"`
	Roles     []userEntity.UserRole `json:"roles"`
	Status    userEntity.UserStatus `json:"status"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type AuthResponse struct {
	User   User      `json:"user"`
	Tokens TokenPair `json:"tokens"`
}

func FromUser(u *userEntity.User, tokens *jwt.TokenPair) AuthResponse {
	return AuthResponse{
		User: User{
			ID:        u.ID.String(),
			TenantID:  u.TenantID.String(),
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			FullName:  u.FullName(),
			Roles:     u.Roles,
			Status:    u.Status,
		},
		Tokens: TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpiresIn:    tokens.ExpiresIn,
			TokenType:    tokens.TokenType,
		},
	}
}
