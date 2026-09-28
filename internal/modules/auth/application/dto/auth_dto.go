package dto

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/jwt"
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

type AuthResponse struct {
	User         UserResponse   `json:"user"`
	TokenPair    jwt.TokenPair  `json:"tokens"`
}

type UserResponse struct {
	ID        string             `json:"id"`
	TenantID  string             `json:"tenant_id"`
	CompanyID *string            `json:"company_id,omitempty"`
	Email     string             `json:"email"`
	FirstName string             `json:"first_name"`
	LastName  string             `json:"last_name"`
	FullName  string             `json:"full_name"`
	Roles     []entity.UserRole  `json:"roles"`
	Status    entity.UserStatus  `json:"status"`
}

func ToAuthResponse(user *entity.User, tokens *jwt.TokenPair) AuthResponse {
	var companyID *string
	if user.CompanyID != nil {
		id := user.CompanyID.String()
		companyID = &id
	}

	return AuthResponse{
		User: UserResponse{
			ID:        user.ID.String(),
			TenantID:  user.TenantID.String(),
			CompanyID: companyID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			FullName:  user.FullName(),
			Roles:     user.Roles,
			Status:    user.Status,
		},
		TokenPair: *tokens,
	}
}

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func ToTokenPairResponse(tokens *jwt.TokenPair) TokenPairResponse {
	return TokenPairResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
	}
}