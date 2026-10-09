package dto

import (
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type User struct {
	ID          string                `json:"id"`
	CompanyID   *string               `json:"company_id,omitempty"`
	RoleID      string                `json:"role_id"`
	Role        string                `json:"role"`
	Permissions []string              `json:"permissions"`
	FirstName   string                `json:"first_name"`
	LastName    string                `json:"last_name"`
	Username    string                `json:"username"`
	Email       string                `json:"email"`
	Status      userEntity.UserStatus `json:"status"`
}

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type AuthResponse struct {
	User  User  `json:"user"`
	Token Token `json:"token"`
}

func FromUser(u *userEntity.AppUser, permissions []string) User {
	var companyID *string
	if u.CompanyID != nil {
		s := u.CompanyID.String()
		companyID = &s
	}

	if permissions == nil {
		permissions = []string{}
	}

	return User{
		ID:          u.ID.String(),
		CompanyID:   companyID,
		RoleID:      u.RoleID.String(),
		Role:        u.RoleName,
		Permissions: permissions,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Username:    u.Username,
		Email:       u.Email,
		Status:      u.Status,
	}
}
