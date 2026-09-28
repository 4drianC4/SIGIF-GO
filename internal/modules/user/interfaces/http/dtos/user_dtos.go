package dtos

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type UserResponse struct {
	ID         uuid.UUID         `json:"id"`
	TenantID   uuid.UUID         `json:"tenant_id"`
	CompanyID  *uuid.UUID        `json:"company_id,omitempty"`
	Email      string            `json:"email"`
	FirstName  string            `json:"first_name"`
	LastName   string            `json:"last_name"`
	FullName   string            `json:"full_name"`
	Phone      string            `json:"phone"`
	AvatarURL  string            `json:"avatar_url"`
	Roles      []entity.UserRole `json:"roles"`
	Status     entity.UserStatus `json:"status"`
	LastLoginAt *string          `json:"last_login_at,omitempty"`
	Settings   entity.UserSettings `json:"settings"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}

func ToUserResponse(u *entity.User) UserResponse {
	var lastLogin *string
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		lastLogin = &s
	}

	return UserResponse{
		ID:         u.ID,
		TenantID:   u.TenantID,
		CompanyID:  u.CompanyID,
		Email:      u.Email,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		FullName:   u.FullName(),
		Phone:      u.Phone,
		AvatarURL:  u.AvatarURL,
		Roles:      u.Roles,
		Status:     u.Status,
		LastLoginAt: lastLogin,
		Settings:   u.Settings,
		CreatedAt:  u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToUserResponseList(users []*entity.User) []UserResponse {
	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = ToUserResponse(u)
	}
	return result
}

type CreateUserRequest struct {
	Email     string            `json:"email" validate:"required,email"`
	Password  string            `json:"password" validate:"required,min=8"`
	FirstName string            `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string            `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string            `json:"phone" validate:"omitempty,max=50"`
	Roles     []entity.UserRole `json:"roles" validate:"omitempty,dive,oneof=super_admin tenant_admin company_admin manager cashier inventory sales viewer"`
	CompanyID *uuid.UUID        `json:"company_id,omitempty"`
}

type UpdateUserRequest struct {
	FirstName string             `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string             `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string             `json:"phone" validate:"omitempty,max=50"`
	AvatarURL string             `json:"avatar_url" validate:"omitempty,url"`
	Roles     []entity.UserRole  `json:"roles" validate:"omitempty,dive,oneof=super_admin tenant_admin company_admin manager cashier inventory sales viewer"`
	Settings  entity.UserSettings `json:"settings"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}