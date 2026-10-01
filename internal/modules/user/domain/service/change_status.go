package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func (s *UserService) Activate(ctx context.Context, id uuid.UUID) error {
	return s.updateStatus(ctx, id, func(u *entity.User) { u.Activate(s.clock) })
}

func (s *UserService) Deactivate(ctx context.Context, id uuid.UUID) error {
	return s.updateStatus(ctx, id, func(u *entity.User) { u.Deactivate(s.clock) })
}

func (s *UserService) Suspend(ctx context.Context, id uuid.UUID) error {
	return s.updateStatus(ctx, id, func(u *entity.User) { u.Suspend(s.clock) })
}

func (s *UserService) updateStatus(ctx context.Context, id uuid.UUID, fn func(*entity.User)) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}

	fn(user)
	return s.repo.Update(ctx, user)
}
