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
	UserID        uuid.UUID
	Reason        string
}

func (s *CatalogService) UpdateProduct(ctx context.Context, params UpdateProductParams) (*entity.Product, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	var product *entity.Product
	err := s.products.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		product, err = s.updateProduct(txCtx, params)
		return err
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *CatalogService) updateProduct(ctx context.Context, params UpdateProductParams) (*entity.Product, error) {
	product, err := s.products.GetByID(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	previousCost := product.CostPrice
	previousPrice := product.SalePrice

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

	priceChanged := !previousPrice.Equal(params.SalePrice)
	costChanged := !previousCost.Equal(params.CostPrice)
	if err := s.products.Update(ctx, product); err != nil {
		return nil, err
	}
	if (priceChanged || costChanged) && params.UserID != uuid.Nil {
		err = s.priceHistory.Record(ctx,
			product.ID, params.UserID,
			previousCost, params.CostPrice,
			previousPrice, params.SalePrice,
			params.Reason,
		)
		if err != nil {
			return nil, err
		}
	}

	return product, nil
}

// GetPriceHistory devuelve el historial de precios paginado de un producto.
type GetPriceHistoryParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	Page      int
	Limit     int
}

func (s *CatalogService) GetPriceHistory(ctx context.Context, params GetPriceHistoryParams) ([]*entity.PriceHistoryEntry, int64, error) {
	if params.CompanyID == uuid.Nil {
		return nil, 0, ErrCompanyRequired
	}
	product, err := s.products.GetByID(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, 0, err
	}
	if product == nil {
		return nil, 0, ErrProductNotFound
	}
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	return s.priceHistory.GetByProduct(ctx, params.ProductID, params.Page, params.Limit)
}
