package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleGetProduct(ctx context.Context, cmd command.GetProduct) (*entity.Product, error) {
	return h.service.GetProduct(ctx, service.GetProductParams{
		CompanyID: cmd.CompanyID,
		ProductID: cmd.ProductID,
	})
}

func (h *CatalogCommandHandler) HandleListProducts(ctx context.Context, cmd command.ListProducts) ([]*entity.Product, int64, error) {
	return h.service.ListProducts(ctx, service.ListProductsParams{
		CompanyID:  cmd.CompanyID,
		Name:       cmd.Name,
		CategoryID: cmd.CategoryID,
		Status:     cmd.Status,
		Page:       cmd.Page,
		Limit:      cmd.Limit,
	})
}

func (h *CatalogCommandHandler) HandleSetProductStatus(ctx context.Context, cmd command.SetProductStatus) (*entity.Product, error) {
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
