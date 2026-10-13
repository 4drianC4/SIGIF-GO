package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type UpdateProductPriceParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	UserID    uuid.UUID
	CostPrice *decimal.Decimal
	SalePrice *decimal.Decimal
	Reason    string
}

func (s *CatalogService) UpdateProductPrice(ctx context.Context, params UpdateProductPriceParams) (*entity.Product, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	if params.UserID == uuid.Nil {
		return nil, sharedErrors.ErrUnauthorized
	}
	if params.CostPrice == nil && params.SalePrice == nil {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "validation failed", 400).
			WithDetails(map[string]string{"prices": "debe enviar cost_price o sale_price"})
	}
	params.Reason = strings.TrimSpace(params.Reason)
	if utf8.RuneCountInString(params.Reason) > 160 {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "validation failed", 400).
			WithDetails(map[string]string{"reason": "admite como máximo 160 caracteres"})
	}

	var product *entity.Product
	err := s.products.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		product, err = s.products.GetByID(txCtx, params.CompanyID, params.ProductID)
		if err != nil {
			return err
		}
		if product == nil {
			return ErrProductNotFound
		}
		previousCost, previousPrice := product.CostPrice, product.SalePrice
		previousUpdatedAt := product.UpdatedAt
		costPrice, salePrice := previousCost, previousPrice
		if params.CostPrice != nil {
			costPrice = *params.CostPrice
		}
		if params.SalePrice != nil {
			salePrice = *params.SalePrice
		}
		if err := product.UpdatePrices(s.clock, costPrice, salePrice); err != nil {
			return err
		}
		if previousCost.Equal(costPrice) && previousPrice.Equal(salePrice) {
			product.UpdatedAt = previousUpdatedAt
			return nil
		}
		if err := s.products.Update(txCtx, product); err != nil {
			return err
		}
		return s.priceHistory.Record(txCtx, product.ID, params.UserID,
			previousCost, costPrice, previousPrice, salePrice, params.Reason)
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}
