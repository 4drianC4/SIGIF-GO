package service

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

// UserService exposes the use cases for application users and RBAC.
type UserService struct {
	repo     repository.UserRepository
	roleRepo repository.RoleRepository
	permRepo repository.PermissionRepository
	clock    clock.Clock
}

func NewUserService(
	repo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permRepo repository.PermissionRepository,
	clock clock.Clock,
) *UserService {
	return &UserService{
		repo:     repo,
		roleRepo: roleRepo,
		permRepo: permRepo,
		clock:    clock,
	}
}
