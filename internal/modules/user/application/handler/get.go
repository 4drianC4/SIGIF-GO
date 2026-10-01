package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

func (h *UserQueryHandler) HandleGet(ctx context.Context, q query.GetUser) (*entity.User, error) {
	return h.service.GetByID(ctx, q.ID)
}

func (h *UserQueryHandler) HandleGetByEmail(ctx context.Context, q query.GetUserByEmail) (*entity.User, error) {
	return h.service.GetByEmail(ctx, q.TenantID, q.Email)
}

func (h *UserQueryHandler) HandleList(ctx context.Context, q query.ListUsers) ([]*entity.User, int64, error) {
	return h.service.List(ctx, q.TenantID, q.Offset, q.Limit)
}
