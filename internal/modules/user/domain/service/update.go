package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type UpdateUserInput struct {
	FirstName string
	LastName  string
	Phone     string
	Area      string
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, in UpdateUserInput) (*entity.AppUser, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}

	user.Update(in.FirstName, in.LastName, in.Phone, in.Area, s.clock)
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
