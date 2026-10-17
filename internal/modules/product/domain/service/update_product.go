package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
)

type UpdateProductParams struct {
	CompanyID       uuid.UUID
	ProductID       uuid.UUID
	CategoryID      uuid.UUID
	UnitOfMeasureID uuid.UUID
	SKU             string
	Barcode         string
	Name            string
	Description     string
	Cost            decimal.Decimal
	SalePrice       decimal.Decimal
	MinStock        *decimal.Decimal
}

func (s *CatalogService) UpdateProduct(ctx context.Context, params UpdateProductParams) (*repository.ProductListItem, error) {
	product, err := s.findProduct(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}

	minStock := product.MinStock
	if params.MinStock != nil {
		minStock = *params.MinStock
	}
	if err := product.Update(s.clock, entity.UpdateProductParams{
		CategoryID:      params.CategoryID,
		UnitOfMeasureID: params.UnitOfMeasureID,
		SKU:             params.SKU,
		Barcode:         params.Barcode,
		Name:            params.Name,
		Description:     params.Description,
		Cost:            params.Cost,
		SalePrice:       params.SalePrice,
		MinStock:        minStock,
	}); err != nil {
		return nil, err
	}

	category, err := s.checkProductReferences(ctx, product, product.ID)
	if err != nil {
		return nil, err
	}

	if err := s.products.Update(ctx, product); err != nil {
		return nil, err
	}
	return &repository.ProductListItem{Product: product, CategoryName: category.Name}, nil
}
