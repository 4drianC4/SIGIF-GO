package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

type CatalogCommandHandler struct {
	service *service.CatalogService
}

func NewCatalogCommandHandler(service *service.CatalogService) *CatalogCommandHandler {
	return &CatalogCommandHandler{service: service}
}

func (h *CatalogCommandHandler) HandleGetPriceHistory(ctx context.Context, companyID, productID uuid.UUID, page, limit int) ([]*entity.PriceHistoryEntry, int64, error) {
	return h.service.GetPriceHistory(ctx, service.GetPriceHistoryParams{
		CompanyID: companyID,
		ProductID: productID,
		Page:      page,
		Limit:     limit,
	})
}
