package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleCreateProduct(ctx context.Context, cmd command.CreateProduct) (*repository.ProductListItem, error) {
	return h.service.CreateProduct(ctx, service.CreateProductParams{
		CompanyID:       cmd.CompanyID,
		CategoryID:      cmd.CategoryID,
		UnitOfMeasureID: cmd.UnitOfMeasureID,
		SKU:             cmd.SKU,
		Barcode:         cmd.Barcode,
		Name:            cmd.Name,
		Description:     cmd.Description,
		Cost:            cmd.Cost,
		SalePrice:       cmd.SalePrice,
		InitialStock:    cmd.InitialStock,
		MinStock:        cmd.MinStock,
	})
}
