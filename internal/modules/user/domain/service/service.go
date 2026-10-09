package service

import (
	companyRepository "github.com/sigif/sigif-go/internal/modules/company/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

// UserService exposes the use cases for application users and RBAC.
type UserService struct {
	companyRepo companyRepository.CompanyRepository
	repo        repository.UserRepository
	roleRepo    repository.RoleRepository
	permRepo    repository.PermissionRepository
	clock       clock.Clock
}

func NewUserService(
	repo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permRepo repository.PermissionRepository,
	clock clock.Clock,
	companyRepo companyRepository.CompanyRepository,
) *UserService {
	return &UserService{
		companyRepo: companyRepo,
		repo:        repo,
		roleRepo:    roleRepo,
		permRepo:    permRepo,
		clock:       clock,
	}
}
