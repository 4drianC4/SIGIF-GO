package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/security"
)

type CreateUserParams struct {
	TenantID  uuid.UUID
	Email     string
	Password  string
	FirstName string
	LastName  string
	Roles     []entity.UserRole
}

// Create crea un nuevo usuario validando unicidad de email por tenant.
func (s *UserService) Create(ctx context.Context, params CreateUserParams) (*entity.User, error) {
	exists, err := s.repo.ExistsByEmail(ctx, params.TenantID, params.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, sharedErrors.New(sharedErrors.CodeConflict, "user with this email already exists", 409)
	}

	passwordHash, err := security.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	user := entity.NewUser(s.clock, params.TenantID, params.Email, passwordHash, params.FirstName, params.LastName, params.Roles)
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
