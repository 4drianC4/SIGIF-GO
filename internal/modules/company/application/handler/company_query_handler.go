package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/application/port"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/company/domain/service"
)

type CompanyQueryHandler struct {
	service *service.CompanyService
}

func NewCompanyQueryHandler(service *service.CompanyService) *CompanyQueryHandler {
	return &CompanyQueryHandler{service: service}
}

func (h *CompanyQueryHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	return h.service.GetByID(ctx, id)
}

func (h *CompanyQueryHandler) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Company, error) {
	return h.service.GetByTenantID(ctx, tenantID)
}

func (h *CompanyQueryHandler) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.Company, int64, error) {
	return h.service.List(ctx, tenantID, offset, limit)
}

var _ port.CompanyQueryPort = (*CompanyQueryHandler)(nil)