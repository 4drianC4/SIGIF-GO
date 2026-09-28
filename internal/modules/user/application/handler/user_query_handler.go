package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/application/port"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
)

type UserQueryHandler struct {
	service *service.UserService
}

func NewUserQueryHandler(service *service.UserService) *UserQueryHandler {
	return &UserQueryHandler{service: service}
}

func (h *UserQueryHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return h.service.GetByID(ctx, id)
}

func (h *UserQueryHandler) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return h.service.GetByEmail(ctx, email)
}

func (h *UserQueryHandler) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error) {
	return h.service.GetByTenantID(ctx, tenantID)
}

func (h *UserQueryHandler) GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*entity.User, error) {
	return h.service.GetByCompanyID(ctx, companyID)
}

func (h *UserQueryHandler) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.User, int64, error) {
	return h.service.List(ctx, tenantID, offset, limit)
}

var _ port.UserQueryPort = (*UserQueryHandler)(nil)