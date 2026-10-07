package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type UpdateProductParams struct {
	CompanyID     uuid.UUID
	ProductID     uuid.UUID
	CategoryID    uuid.UUID
	SKU           string
	Barcode       string
	Name          string
	Description   string
	UnitOfMeasure entity.UnitOfMeasure
	CostPrice     decimal.Decimal
	SalePrice     decimal.Decimal
}

func (s *CatalogService) UpdateProduct(ctx context.Context, params UpdateProductParams) (*entity.Product, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}

	product, err := s.products.GetByID(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	if err := product.Update(s.clock, entity.UpdateProductParams{
		CategoryID:    params.CategoryID,
		SKU:           params.SKU,
		Barcode:       params.Barcode,
		Name:          params.Name,
		Description:   params.Description,
		UnitOfMeasure: params.UnitOfMeasure,
		CostPrice:     params.CostPrice,
		SalePrice:     params.SalePrice,
	}); err != nil {
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

	skuTaken, err := s.products.ExistsBySKUExcluding(ctx, product.CompanyID, product.SKU, product.ID)
	if err != nil {
		return nil, err
	}
	if skuTaken {
		return nil, ErrSKUTaken
	}

	if product.Barcode != "" {
		barcodeTaken, err := s.products.ExistsByBarcodeExcluding(ctx, product.CompanyID, product.Barcode, product.ID)
		if err != nil {
			return nil, err
		}
		if barcodeTaken {
			return nil, ErrBarcodeTaken
		}
	}

	if err := s.products.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}
