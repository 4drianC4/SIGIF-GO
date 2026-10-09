package dtos

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/application/dto"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type RegisterUserRequest struct {
	FirstName string    `json:"first_name" validate:"required,min=1,max=80"`
	LastName  string    `json:"last_name" validate:"required,min=1,max=80"`
	Email     string    `json:"email" validate:"required,email"`
	CompanyID uuid.UUID `json:"company_id" validate:"required"`
	Role      string    `json:"role" validate:"required,oneof=business_admin employee"`
}

// EditUserRequest carries the fields an administrator may edit on a user: the
// registration fields plus an optional password reset. Every field is optional; omitted fields are
// left unchanged (JSON merge-patch semantics: send only what changes).
type EditUserRequest struct {
	CompanyID *uuid.UUID `json:"company_id" validate:"omitempty"`
	FirstName *string    `json:"first_name" validate:"omitempty,min=1,max=80"`
	LastName  *string    `json:"last_name" validate:"omitempty,min=1,max=80"`
	Email     *string    `json:"email" validate:"omitempty,email"`
	Password  *string    `json:"password" validate:"omitempty,min=8"`
	Role      *string    `json:"role" validate:"omitempty,oneof=business_admin employee"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type UserResponse = dto.User

func ToResponse(u *entity.AppUser) dto.User {
	return dto.FromEntity(u)
}

func ToResponseList(users []*entity.AppUser) []dto.User {
	return dto.FromEntityList(users)
}

// GeneratedPassword is returned only when creating the account.
type RegisterUserResponse struct {
	dto.User
	GeneratedPassword string `json:"generated_password"`
}

// CreatePermissionRequest is the body of POST /permissions (HU-082-01). The
// code is optional and, when sent, must match module.operation.
type CreatePermissionRequest struct {
	Module      string `json:"module" validate:"required,min=1,max=60"`
	Operation   string `json:"operation" validate:"required,min=1,max=60"`
	Code        string `json:"code" validate:"omitempty,max=80"`
	Description string `json:"description" validate:"omitempty,max=200"`
}

// ListPermissionsRequest holds the query string of GET /permissions. Page and
// limit are pointers so an explicit 0 is rejected instead of defaulted.
type ListPermissionsRequest struct {
	Module string `json:"module" query:"module" validate:"omitempty,max=60"`
	Page   *int   `json:"page" query:"page" validate:"omitempty,min=1"`
	Limit  *int   `json:"limit" query:"limit" validate:"omitempty,min=1"`
}
