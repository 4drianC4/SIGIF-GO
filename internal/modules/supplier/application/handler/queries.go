package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/query"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
)

func (h *SupplierQueryHandler) HandleList(ctx context.Context, q query.ListSuppliers) ([]repository.SupplierListItem, int64, error) {
	return h.service.List(ctx, q.CompanyID, q.Filter, q.Offset, q.Limit)
}

func (h *SupplierQueryHandler) HandleListProductSuppliers(ctx context.Context, q query.ListProductSuppliers) ([]repository.SupplierItem, int64, error) {
	return h.service.ListProductSuppliers(ctx, q.CompanyID, q.ProductID, q.Search, q.Offset, q.Limit)
}
func (h *SupplierQueryHandler) HandleProductOptions(ctx context.Context, q query.ProductOptions) ([]repository.ProductOption, int64, error) {
	return h.service.ProductOptions(ctx, q.CompanyID, q.SupplierID, q.Filter, q.Offset, q.Limit)
}
func (h *SupplierQueryHandler) HandleSupplierOptions(ctx context.Context, q query.SupplierOptions) ([]repository.SupplierOption, int64, error) {
	return h.service.SupplierOptions(ctx, q.CompanyID, q.ProductID, q.Search, q.Offset, q.Limit)
}
func (h *SupplierQueryHandler) HandleGet(ctx context.Context, q query.GetSupplier) (*entity.Supplier, error) {
	return h.service.Get(ctx, q.CompanyID, q.ID)
}
func (h *SupplierQueryHandler) HandleSummary(ctx context.Context, q query.SupplierSummary) (repository.SupplierSummary, error) {
	return h.service.Summary(ctx, q.CompanyID)
}
func (h *SupplierQueryHandler) HandleListProducts(ctx context.Context, q query.ListSupplierProducts) ([]repository.ProductItem, int64, error) {
	return h.service.ListProducts(ctx, q.CompanyID, q.SupplierID, q.Search, q.Offset, q.Limit)
}
func (h *SupplierQueryHandler) HandleFormOptions(ctx context.Context, companyID uuid.UUID) (repository.FormOptions, error) {
	return h.service.FormOptions(ctx, companyID)
}
