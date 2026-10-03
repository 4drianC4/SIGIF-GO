package service

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

func (s *UserService) List(ctx context.Context, offset, limit int) ([]*entity.AppUser, int64, error) {
	return s.repo.List(ctx, offset, limit)
}
