package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// EditUserInput holds registration fields and an optional password reset.
// Nil leaves a field unchanged.
type EditUserInput struct {
	CompanyID *uuid.UUID
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
	RoleName  *string
}

// Edit applies a partial administrative update to an existing user.
func (s *UserService) Edit(ctx context.Context, id uuid.UUID, in EditUserInput) (*entity.AppUser, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}

	if in.CompanyID != nil {
		if err := s.validateCompany(ctx, in.CompanyID); err != nil {
			return nil, err
		}
	}
	if in.Email != nil && *in.Email != user.Email {
		exists, err := s.repo.ExistsByEmail(ctx, *in.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, sharedErrors.New(sharedErrors.CodeConflict, "user with this email already exists", 409)
		}
	}

	var roleID *uuid.UUID
	if in.RoleName != nil {
		if !entity.IsUserAssignable(*in.RoleName) {
			return nil, sharedErrors.New(sharedErrors.CodeValidation, "invalid role", 400)
		}
		role, err := s.roleRepo.GetByName(ctx, *in.RoleName)
		if err != nil {
			return nil, err
		}
		if role == nil {
			return nil, sharedErrors.New(sharedErrors.CodeValidation, "invalid role", 400)
		}
		if !role.IsActive() {
			return nil, ErrRoleInactive
		}
		roleID = &role.ID
	}

	user.Edit(entity.EditUserParams{
		CompanyID: in.CompanyID,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
		RoleID:    roleID,
	}, s.clock)

	if in.RoleName != nil {
		user.RoleName = *in.RoleName
	}
	if in.Password != nil {
		if err := user.ChangePassword(*in.Password, s.clock); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
