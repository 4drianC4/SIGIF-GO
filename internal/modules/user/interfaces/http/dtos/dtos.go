package dtos

import (
	"github.com/sigif/sigif-go/internal/modules/user/application/dto"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type RegisterUserRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=80"`
	LastName  string `json:"last_name" validate:"required,min=1,max=80"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	Role      string `json:"role" validate:"required,min=1,max=60"`
	Area      string `json:"area" validate:"omitempty,max=120"`
}

// EditUserRequest carries the fields an administrator may edit on a user: the
// same ones used at registration. Every field is optional; omitted fields are
// left unchanged (JSON merge-patch semantics: send only what changes).
type EditUserRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=80"`
	LastName  *string `json:"last_name" validate:"omitempty,min=1,max=80"`
	Email     *string `json:"email" validate:"omitempty,email"`
	Password  *string `json:"password" validate:"omitempty,min=8"`
	Role      *string `json:"role" validate:"omitempty,min=1,max=60"`
	Area      *string `json:"area" validate:"omitempty,max=120"`
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

// StatusAll lists roles in every status.
const StatusAll = "all"

// ListRolesRequest holds the query string of GET /roles (HU-082-02). Page and
// limit are pointers so an explicit 0 is rejected instead of defaulted.
type ListRolesRequest struct {
	Q      string `json:"q" query:"q" validate:"max=100"`
	Type   string `json:"type" query:"type" validate:"omitempty,oneof=system custom"`
	Status string `json:"status" query:"status" validate:"omitempty,oneof=active inactive all"`
	Page   *int   `json:"page" query:"page" validate:"omitempty,min=1"`
	Limit  *int   `json:"limit" query:"limit" validate:"omitempty,min=1"`
}
