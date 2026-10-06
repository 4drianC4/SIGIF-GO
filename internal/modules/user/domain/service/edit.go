package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// EditUserInput holds the data an administrator may edit for a user: the same
// fields that are required/allowed when registering (first_name, last_name,
// email, password, role, area). Nil = the field is left unchanged.
type EditUserInput struct {
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
	RoleName  *string
	Area      *string
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
		role, err := s.roleRepo.GetByName(ctx, *in.RoleName)
		if err != nil {
			return nil, err
		}
		if role == nil {
			return nil, sharedErrors.New(sharedErrors.CodeValidation, "invalid role", 400)
		}
		roleID = &role.ID
	}

	user.Edit(entity.EditUserParams{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
		RoleID:    roleID,
		Area:      in.Area,
	}, s.clock)

	if in.Password != nil && *in.Password != "" {
		if err := user.ChangePassword(*in.Password, s.clock); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
