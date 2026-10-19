package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
)

const noMovementWindow = 90 * 24 * time.Hour

type CreateProductParams struct {
	CompanyID       uuid.UUID
	CategoryID      uuid.UUID
	UnitOfMeasureID uuid.UUID
	SKU             string
	Barcode         string
	Name            string
	Description     string
	Cost            decimal.Decimal
	SalePrice       decimal.Decimal
	InitialStock    decimal.Decimal
	MinStock        decimal.Decimal
}

// CreateProduct registers a product. It runs in a transaction that locks the
// category, so the category cannot be deleted while the product is created.
func (s *CatalogService) CreateProduct(ctx context.Context, params CreateProductParams) (*repository.ProductListItem, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}

	product, err := entity.NewProduct(s.clock, entity.NewProductParams{
		CompanyID:       params.CompanyID,
		CategoryID:      params.CategoryID,
		UnitOfMeasureID: params.UnitOfMeasureID,
		SKU:             params.SKU,
		Barcode:         params.Barcode,
		Name:            params.Name,
		Description:     params.Description,
		Cost:            params.Cost,
		SalePrice:       params.SalePrice,
		InitialStock:    params.InitialStock,
		MinStock:        params.MinStock,
	})
	if err != nil {
		return nil, err
	}

	var item *repository.ProductListItem
	err = s.products.WithinTransaction(ctx, func(txCtx context.Context) error {
		category, err := s.checkProductReferences(txCtx, product, uuid.Nil)
		if err != nil {
			return err
		}
		if err := s.products.Create(txCtx, product); err != nil {
			return err
		}
		item = &repository.ProductListItem{Product: product, CategoryName: category.Name}
		return nil
	})
	if err != nil {
		// A unique index violation aborts the transaction, so the product that
		// won is looked up outside of it.
		return nil, s.detailDuplicate(ctx, err, product.CompanyID, product.Name, product.SKU, product.Barcode, uuid.Nil)
	}
	return item, nil
}

type ListProductsParams struct {
	CompanyID uuid.UUID
	Filter    repository.ProductFilter
	Offset    int
	Limit     int
}

func (s *CatalogService) ListProducts(ctx context.Context, params ListProductsParams) ([]repository.ProductListItem, int64, error) {
	if params.CompanyID == uuid.Nil {
		return nil, 0, ErrCompanyRequired
	}
	return s.products.List(ctx, params.CompanyID, params.Filter, params.Offset, params.Limit)
}

func (s *CatalogService) ProductSummary(ctx context.Context, companyID uuid.UUID) (repository.ProductSummary, error) {
	if companyID == uuid.Nil {
		return repository.ProductSummary{}, ErrCompanyRequired
	}
	return s.products.Summary(ctx, companyID, s.clock.NowUTC().Add(-noMovementWindow))
}

type ValidateDuplicateParams struct {
	CompanyID uuid.UUID
	Name      string
	SKU       string
	Barcode   string
	ExcludeID uuid.UUID
}

// DuplicateResult tells whether any value is already held and by which
// product. Field is the first duplicated field.
type DuplicateResult struct {
	Exists    bool
	Field     string
	Conflicts []DuplicateConflict
}

func (s *CatalogService) ValidateDuplicate(ctx context.Context, params ValidateDuplicateParams) (*DuplicateResult, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	name, sku := entity.NormalizeName(params.Name), entity.NormalizeSKU(params.SKU)
	if name == "" && sku == "" && params.Barcode == "" {
		return nil, ErrDuplicateCheckFields
	}

	conflicts, err := s.findConflicts(ctx, params.CompanyID, name, sku, strings.TrimSpace(params.Barcode), params.ExcludeID)
	if err != nil {
		return nil, err
	}
	result := &DuplicateResult{Exists: len(conflicts) > 0, Conflicts: conflicts}
	if result.Exists {
		result.Field = conflicts[0].Field
	}
	return result, nil
}

func (s *CatalogService) checkProductReferences(ctx context.Context, product *entity.Product, excludeID uuid.UUID) (*entity.Category, error) {
	category, err := s.categories.GetByID(ctx, product.CompanyID, product.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}
	if !category.IsActive() {
		return nil, ErrCategoryInactive
	}

	unitExists, err := s.units.Exists(ctx, product.UnitOfMeasureID)
	if err != nil {
		return nil, err
	}
	if !unitExists {
		return nil, ErrUnitNotFound
	}

	conflicts, err := s.findConflicts(ctx, product.CompanyID, product.Name, product.SKU, product.Barcode, excludeID)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, newDuplicateError(conflicts)
	}
	return category, nil
}

func (s *CatalogService) findProduct(ctx context.Context, companyID, productID uuid.UUID) (*entity.Product, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	product, err := s.products.GetByID(ctx, companyID, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

func (s *CatalogService) withCategoryName(ctx context.Context, product *entity.Product) (*repository.ProductListItem, error) {
	category, err := s.categories.GetByID(ctx, product.CompanyID, product.CategoryID)
	if err != nil {
		return nil, err
	}
	item := &repository.ProductListItem{Product: product}
	if category != nil {
		item.CategoryName = category.Name
	}
	return item, nil
}
