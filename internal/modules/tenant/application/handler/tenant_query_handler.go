package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/application/port"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/service"
)

type TenantQueryHandler struct {
	service *service.TenantService
}

func NewTenantQueryHandler(service *service.TenantService) *TenantQueryHandler {
	return &TenantQueryHandler{service: service}
}

func (h *TenantQueryHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	return h.service.GetByID(ctx, id)
}

func (h *TenantQueryHandler) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	return h.service.GetBySlug(ctx, slug)
}

func (h *TenantQueryHandler) List(ctx context.Context, offset, limit int) ([]*entity.Tenant, int64, error) {
	return h.service.List(ctx, offset, limit)
}

var _ port.TenantQueryPort = (*TenantQueryHandler)(nil)