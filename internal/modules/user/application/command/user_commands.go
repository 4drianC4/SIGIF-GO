package command

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type CreateUserCommand struct {
	TenantID  uuid.UUID
	CompanyID *uuid.UUID
	Email     string
	Password  string
	FirstName string
	LastName  string
	Roles     []entity.UserRole
}

type UpdateUserCommand struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Phone     string
	AvatarURL string
	Roles     []entity.UserRole
	Settings  entity.UserSettings
}

type ChangePasswordCommand struct {
	ID             uuid.UUID
	CurrentPassword string
	NewPassword     string
}

type DeleteUserCommand struct {
	ID uuid.UUID
}

type ActivateUserCommand struct {
	ID uuid.UUID
}

type DeactivateUserCommand struct {
	ID uuid.UUID
}

type SuspendUserCommand struct {
	ID uuid.UUID
}

type RecordLoginCommand struct {
	ID uuid.UUID
}