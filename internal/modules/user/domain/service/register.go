package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/security"
)

type RegisterUserInput struct {
	CompanyID *uuid.UUID
	RoleName  string
	FirstName string
	LastName  string
	Email     string
}

type RegisteredUser struct {
	User              *entity.AppUser
	GeneratedPassword string
}

// Register creates a new user after validating uniqueness and resolving the
// requested role by name.
func (s *UserService) Register(ctx context.Context, in RegisterUserInput) (*RegisteredUser, error) {
	if err := s.validateCompany(ctx, in.CompanyID); err != nil {
		return nil, err
	}
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
	if !role.IsActive() {
		return nil, ErrRoleInactive
	}

	exists, err := s.repo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, sharedErrors.New(sharedErrors.CodeConflict, "user with this email already exists", 409)
	}

	password, err := security.GenerateToken(24)
	if err != nil {
		return nil, err
	}
	user, err := entity.NewUser(s.clock, entity.RegisterUserParams{
		CompanyID: in.CompanyID,
		RoleID:    role.ID,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
		Password:  password,
	})
	if err != nil {
		return nil, err
	}

	user.RoleName = role.Name
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisteredUser{User: user, GeneratedPassword: password}, nil
}
