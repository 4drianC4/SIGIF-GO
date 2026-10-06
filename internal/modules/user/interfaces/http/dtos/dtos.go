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
