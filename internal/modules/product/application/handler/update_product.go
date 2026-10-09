package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleUpdateProduct(ctx context.Context, cmd command.UpdateProduct) (*repository.ProductListItem, error) {
	return h.service.UpdateProduct(ctx, service.UpdateProductParams{
		CompanyID:       cmd.CompanyID,
		ProductID:       cmd.ProductID,
		CategoryID:      cmd.CategoryID,
		UnitOfMeasureID: cmd.UnitOfMeasureID,
		SKU:             cmd.SKU,
		Barcode:         cmd.Barcode,
		Name:            cmd.Name,
		Description:     cmd.Description,
		Cost:            cmd.Cost,
		SalePrice:       cmd.SalePrice,
		MinStock:        cmd.MinStock,
	})
}

func (h *CatalogCommandHandler) HandleSetProductStatus(ctx context.Context, cmd command.SetProductStatus) (*repository.ProductListItem, error) {
	return h.service.SetProductStatus(ctx, service.SetProductStatusParams{
		CompanyID: cmd.CompanyID,
		ProductID: cmd.ProductID,
		Status:    cmd.Status,
	})
}

func (h *CatalogCommandHandler) HandleDeleteProduct(ctx context.Context, cmd command.DeleteProduct) error {
	return h.service.DeleteProduct(ctx, service.DeleteProductParams{
		CompanyID: cmd.CompanyID,
		ProductID: cmd.ProductID,
	})
}
