package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleUpdateProductPrice(ctx context.Context, cmd command.UpdateProductPrice) (*repository.ProductListItem, error) {
	return h.service.UpdateProductPrice(ctx, service.UpdateProductPriceParams{
		CompanyID: cmd.CompanyID,
		ProductID: cmd.ProductID,
		UserID:    cmd.UserID,
		Cost:      cmd.Cost,
		SalePrice: cmd.SalePrice,
		Reason:    cmd.Reason,
	})
}
