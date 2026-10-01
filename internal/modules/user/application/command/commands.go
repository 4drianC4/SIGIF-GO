package command

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type CreateUser struct {
	TenantID  uuid.UUID
	Email     string
	Password  string
	FirstName string
	LastName  string
	Roles     []entity.UserRole
}

type UpdateUser struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Phone     string
	AvatarURL string
	Roles     []entity.UserRole
	Settings  entity.UserSettings
}

type ChangePassword struct {
	ID              uuid.UUID
	CurrentPassword string
	NewPassword     string
}

type DeleteUser struct {
	ID uuid.UUID
}

type ChangeStatus struct {
	ID     uuid.UUID
	Action entity.UserStatus
}
