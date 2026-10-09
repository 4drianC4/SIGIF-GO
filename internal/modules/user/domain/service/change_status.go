package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func (s *UserService) Activate(ctx context.Context, id uuid.UUID) error {
	return s.changeStatus(ctx, id, entity.UserStatusActive)
}

func (s *UserService) Deactivate(ctx context.Context, id uuid.UUID) error {
	if actorID, ok := middleware.UserIDFromContext(ctx); ok && actorID == id {
		return sharedErrors.New(sharedErrors.CodeConflict, "you cannot deactivate your own account", 409)
	}
	return s.changeStatus(ctx, id, entity.UserStatusInactive)
}

func (s *UserService) changeStatus(ctx context.Context, id uuid.UUID, status entity.UserStatus) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}

	switch status {
	case entity.UserStatusActive:
		user.Activate(s.clock)
	case entity.UserStatusInactive:
		if user.IsDeactivated() {
			return sharedErrors.New(sharedErrors.CodeConflict, "user is already inactive", 409)
		}
		user.Deactivate(s.clock)
	default:
		return sharedErrors.New(sharedErrors.CodeBadRequest, "invalid status", 400)
	}

	return s.repo.Update(ctx, user)
}
