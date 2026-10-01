package service

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

// UserService expone los casos de uso de dominio de usuarios.
// Cada método está en su propio archivo (create.go, get.go, etc.).
type UserService struct {
	repo  repository.UserRepository
	clock clock.Clock
}

func NewUserService(repo repository.UserRepository, clock clock.Clock) *UserService {
	return &UserService{
		repo:  repo,
		clock: clock,
	}
}
