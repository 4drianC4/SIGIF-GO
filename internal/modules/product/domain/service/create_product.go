package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type CreateProductParams struct {
	CompanyID     uuid.UUID
	CategoryID    uuid.UUID
	SKU           string
	Barcode       string
	Name          string
	Description   string
	UnitOfMeasure entity.UnitOfMeasure
	CostPrice     decimal.Decimal
	SalePrice     decimal.Decimal
}

func (s *CatalogService) CreateProduct(ctx context.Context, params CreateProductParams) (*entity.Product, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}

	product, err := entity.NewProduct(s.clock, entity.NewProductParams{
		CompanyID:     params.CompanyID,
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

	skuTaken, err := s.products.ExistsBySKU(ctx, product.CompanyID, product.SKU)
	if err != nil {
		return nil, err
	}
	if skuTaken {
		return nil, ErrSKUTaken
	}

	if product.Barcode != "" {
		barcodeTaken, err := s.products.ExistsByBarcode(ctx, product.CompanyID, product.Barcode)
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
