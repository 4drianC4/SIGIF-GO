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
	UserID          uuid.UUID
	Reason          string
}

// UpdateProduct edits a product and, when its cost or selling price change and
// the author is known, records the change in the price history. Both writes
// share one transaction and the product row is locked while it is read.
func (s *CatalogService) UpdateProduct(ctx context.Context, params UpdateProductParams) (*repository.ProductListItem, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	var item *repository.ProductListItem
	err := s.products.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		item, err = s.updateProduct(txCtx, params)
		return err
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *CatalogService) updateProduct(ctx context.Context, params UpdateProductParams) (*repository.ProductListItem, error) {
	product, err := s.findProduct(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}

	previousCost, previousPrice := product.Cost, product.SalePrice
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

	pricesChanged := !previousCost.Equal(product.Cost) || !previousPrice.Equal(product.SalePrice)
	if pricesChanged && params.UserID != uuid.Nil {
		if err := s.priceHistory.Record(ctx, product.ID, params.UserID,
			previousCost, product.Cost, previousPrice, product.SalePrice, params.Reason); err != nil {
			return nil, err
		}
	}
	return &repository.ProductListItem{Product: product, CategoryName: category.Name}, nil
}

type GetPriceHistoryParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	Page      int
	Limit     int
}

// GetPriceHistory returns the paginated price changes of a product, newest first.
func (s *CatalogService) GetPriceHistory(ctx context.Context, params GetPriceHistoryParams) ([]*entity.PriceHistoryEntry, int64, error) {
	if _, err := s.findProduct(ctx, params.CompanyID, params.ProductID); err != nil {
		return nil, 0, err
	}
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	return s.priceHistory.GetByProduct(ctx, params.ProductID, params.Page, params.Limit)
}
