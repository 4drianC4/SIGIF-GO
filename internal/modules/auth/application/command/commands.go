package command

import (
	"github.com/google/uuid"
)

type Login struct {
	TenantID  uuid.UUID
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type Refresh struct {
	RefreshToken string
}

type Logout struct {
	RefreshToken string
}

type LogoutAll struct {
	UserID uuid.UUID
}
