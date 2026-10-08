package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type RegisterUserInput struct {
	CompanyID *uuid.UUID
	RoleName  string
	FirstName string
	LastName  string
	Email     string
	Password  string
	Area      string
}

// Register creates a new user after validating uniqueness and resolving the
// requested role by name.
func (s *UserService) Register(ctx context.Context, in RegisterUserInput) (*entity.AppUser, error) {
	if !entity.IsUserAssignable(in.RoleName) {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "invalid role", 400)
	}

	role, err := s.roleRepo.GetByName(ctx, in.RoleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "invalid role", 400)
	}

	exists, err := s.repo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, sharedErrors.New(sharedErrors.CodeConflict, "user with this email already exists", 409)
	}

	user, err := entity.NewUser(s.clock, entity.RegisterUserParams{
		CompanyID: in.CompanyID,
		RoleID:    role.ID,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
		Password:  in.Password,
		Area:      in.Area,
	})
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
