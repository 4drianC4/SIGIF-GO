package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleUpdateProduct(ctx context.Context, cmd command.UpdateProduct) (*entity.Product, error) {
	return h.service.UpdateProduct(ctx, service.UpdateProductParams{
		CompanyID:     cmd.CompanyID,
		ProductID:     cmd.ProductID,
		CategoryID:    cmd.CategoryID,
		SKU:           cmd.SKU,
		Barcode:       cmd.Barcode,
		Name:          cmd.Name,
		Description:   cmd.Description,
		UnitOfMeasure: cmd.UnitOfMeasure,
		CostPrice:     cmd.CostPrice,
		SalePrice:     cmd.SalePrice,
	})
}
