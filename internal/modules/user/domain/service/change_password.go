package service

import (
	"context"

	"github.com/google/uuid"

	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/security"
)

func (s *UserService) ChangePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}

	if err := security.VerifyPassword(currentPassword, user.PasswordHash); err != nil {
		return sharedErrors.New(sharedErrors.CodeUnauthorized, "current password is incorrect", 401)
	}

	if err := user.ChangePassword(newPassword, s.clock); err != nil {
		return err
	}

	return s.repo.Update(ctx, user)
}
