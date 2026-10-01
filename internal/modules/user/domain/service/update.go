package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type UpdateUserParams struct {
	FirstName string
	LastName  string
	Phone     string
	AvatarURL string
	Roles     []entity.UserRole
	Settings  entity.UserSettings
}

// Update actualiza los datos de un usuario existente.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, params UpdateUserParams) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}

	user.Update(params.FirstName, params.LastName, params.Phone, params.AvatarURL, params.Roles, params.Settings, s.clock)
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
