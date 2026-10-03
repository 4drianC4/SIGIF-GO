package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type CreateProductParams struct {
	TenantID      uuid.UUID
	CategoryID    uuid.UUID
	SKU           string
	Barcode       string
	Name          string
	Description   string
	UnitOfMeasure entity.UnitOfMeasure
	CostPrice     decimal.Decimal
	SalePrice     decimal.Decimal
}

// CreateProduct registra un producto activo. Antes de persistir comprueba que la
// categoría exista y esté activa, y que el SKU y el código de barras no se repitan en el tenant.
func (s *CatalogService) CreateProduct(ctx context.Context, params CreateProductParams) (*entity.Product, error) {
	if params.TenantID == uuid.Nil {
		return nil, sharedErrors.ErrTenantRequired
	}

	product, err := entity.NewProduct(s.clock, entity.NewProductParams{
		TenantID:      params.TenantID,
		CategoryID:    params.CategoryID,
		SKU:           params.SKU,
		Barcode:       params.Barcode,
		Name:          params.Name,
		Description:   params.Description,
		UnitOfMeasure: params.UnitOfMeasure,
		CostPrice:     params.CostPrice,
		SalePrice:     params.SalePrice,
	})
	if err != nil {
		return nil, err
	}

	category, err := s.categories.GetByID(ctx, product.TenantID, product.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}
	if !category.IsActive() {
		return nil, ErrCategoryInactive
	}

	skuTaken, err := s.products.ExistsBySKU(ctx, product.TenantID, product.SKU)
	if err != nil {
		return nil, err
	}
	if skuTaken {
		return nil, ErrSKUTaken
	}

	if product.Barcode != "" {
		barcodeTaken, err := s.products.ExistsByBarcode(ctx, product.TenantID, product.Barcode)
		if err != nil {
			return nil, err
		}
		if barcodeTaken {
			return nil, ErrBarcodeTaken
		}
	}

	if err := s.products.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}
