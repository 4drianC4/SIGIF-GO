package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// GetByID obtiene un usuario por ID o error 404.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}
	return user, nil
}

// GetByEmail obtiene un usuario por email dentro de un tenant, o nil si no existe.
func (s *UserService) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error) {
	return s.repo.GetByEmail(ctx, tenantID, email)
}
