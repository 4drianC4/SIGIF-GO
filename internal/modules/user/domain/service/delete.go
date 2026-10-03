package service

import (
	"context"

	"github.com/google/uuid"

	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}

	user.SoftDelete(s.clock)
	return s.repo.Update(ctx, user)
}
