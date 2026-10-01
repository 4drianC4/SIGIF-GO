package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// List retorna usuarios paginados de un tenant.
func (s *UserService) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.User, int64, error) {
	return s.repo.List(ctx, tenantID, offset, limit)
}
