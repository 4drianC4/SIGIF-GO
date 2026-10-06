package service

import (
	"context"
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

	field, err := s.duplicateField(ctx, product.CompanyID, product.Name, product.SKU, product.Barcode)
	if err != nil {
		return nil, err
	}
	if field != "" {
		return nil, duplicateErrors[field]
	}

	if err := s.products.Create(ctx, product); err != nil {
		return nil, err
	}
	return &repository.ProductListItem{Product: product, CategoryName: category.Name}, nil
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

type DuplicateResult struct {
	Exists bool
	Field  string
}

func (s *CatalogService) ValidateDuplicate(ctx context.Context, companyID uuid.UUID, name, sku, barcode string) (*DuplicateResult, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	name, sku = entity.NormalizeName(name), entity.NormalizeSKU(sku)
	if name == "" && sku == "" && barcode == "" {
		return nil, ErrDuplicateCheckFields
	}

	field, err := s.duplicateField(ctx, companyID, name, sku, barcode)
	if err != nil {
		return nil, err
	}
	return &DuplicateResult{Exists: field != "", Field: field}, nil
}

var duplicateErrors = map[string]error{
	"name":    ErrProductNameTaken,
	"sku":     ErrSKUTaken,
	"barcode": ErrBarcodeTaken,
}

func (s *CatalogService) duplicateField(ctx context.Context, companyID uuid.UUID, name, sku, barcode string) (string, error) {
	checks := []struct {
		field  string
		value  string
		exists func(context.Context, uuid.UUID, string) (bool, error)
	}{
		{"name", name, s.products.ExistsByName},
		{"sku", sku, s.products.ExistsBySKU},
		{"barcode", barcode, s.products.ExistsByBarcode},
	}
	for _, check := range checks {
		if check.value == "" {
			continue
		}
		taken, err := check.exists(ctx, companyID, check.value)
		if err != nil {
			return "", err
		}
		if taken {
			return check.field, nil
		}
	}
	return "", nil
}
