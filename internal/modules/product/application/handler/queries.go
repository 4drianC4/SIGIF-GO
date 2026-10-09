package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/application/query"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogQueryHandler) HandleListProducts(ctx context.Context, q query.ListProducts) ([]repository.ProductListItem, int64, error) {
	return h.service.ListProducts(ctx, service.ListProductsParams{
		CompanyID: q.CompanyID,
		Filter: repository.ProductFilter{
			Search:     q.Search,
			CategoryID: q.CategoryID,
			Status:     q.Status,
		},
		Offset: q.Offset,
		Limit:  q.Limit,
	})
}

func (h *CatalogQueryHandler) HandleProductSummary(ctx context.Context, companyID uuid.UUID) (repository.ProductSummary, error) {
	return h.service.ProductSummary(ctx, companyID)
}

func (h *CatalogQueryHandler) HandleGetProduct(ctx context.Context, q query.GetProduct) (*repository.ProductListItem, error) {
	return h.service.GetProduct(ctx, service.GetProductParams{
		CompanyID: q.CompanyID,
		ProductID: q.ProductID,
	})
}

func (h *CatalogQueryHandler) HandleValidateDuplicate(ctx context.Context, q query.ValidateDuplicate) (*service.DuplicateResult, error) {
	return h.service.ValidateDuplicate(ctx, service.ValidateDuplicateParams{
		CompanyID: q.CompanyID,
		Name:      q.Name,
		SKU:       q.SKU,
		Barcode:   q.Barcode,
		ExcludeID: q.ExcludeID,
	})
}

func (h *CatalogQueryHandler) HandleListCategories(ctx context.Context, companyID uuid.UUID) ([]repository.CategoryListItem, error) {
	return h.service.ListCategories(ctx, companyID)
}

func (h *CatalogQueryHandler) HandleListUnitsOfMeasure(ctx context.Context) ([]*entity.UnitOfMeasure, error) {
	return h.service.ListUnitsOfMeasure(ctx)
}

func (h *CatalogQueryHandler) HandleListTaxes(ctx context.Context) ([]*entity.Tax, error) {
	return h.service.ListTaxes(ctx)
}
