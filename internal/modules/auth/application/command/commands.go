package command

import (
	"github.com/google/uuid"
)

type Login struct {
	Email     string
	Password  string
	IPAddress string
	Device    string
}

type Logout struct {
	TokenHash string
}

type Me struct {
	UserID uuid.UUID
}
