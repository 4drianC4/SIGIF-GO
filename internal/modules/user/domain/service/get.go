package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*entity.AppUser, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*entity.AppUser, error) {
	return s.repo.GetByEmail(ctx, email)
}
