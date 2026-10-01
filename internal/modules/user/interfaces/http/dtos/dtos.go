package dtos

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/dto"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type UserResponse struct {
	ID          string             `json:"id"`
	TenantID    string             `json:"tenant_id"`
	Email       string             `json:"email"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	FullName    string             `json:"full_name"`
	Phone       string             `json:"phone,omitempty"`
	AvatarURL   string             `json:"avatar_url,omitempty"`
	Roles       []entity.UserRole  `json:"roles"`
	Status      entity.UserStatus  `json:"status"`
	LastLoginAt *string            `json:"last_login_at,omitempty"`
	Settings    entity.UserSettings `json:"settings"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

type CreateUserRequest struct {
	Email     string            `json:"email" validate:"required,email"`
	Password  string            `json:"password" validate:"required,min=8"`
	FirstName string            `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string            `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string            `json:"phone" validate:"omitempty,max=50"`
	Roles     []entity.UserRole `json:"roles" validate:"omitempty,dive,oneof=super_admin tenant_admin company_admin manager cashier inventory sales viewer"`
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

func toResponse(u dto.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		TenantID:    u.TenantID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		FullName:    u.FullName,
		Phone:       u.Phone,
		AvatarURL:   u.AvatarURL,
		Roles:       u.Roles,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		Settings:    u.Settings,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func ToResponse(u *entity.User) UserResponse {
	return toResponse(dto.FromEntity(u))
}

func ToResponseList(users []*entity.User) []UserResponse {
	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = toResponse(dto.FromEntity(u))
	}
	return result
}

func TenantIDFromContext(c *fiber.Ctx) uuid.UUID {
	v := c.Locals("tenant_id")
	if v == nil {
		return uuid.Nil
	}
	if id, ok := v.(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}