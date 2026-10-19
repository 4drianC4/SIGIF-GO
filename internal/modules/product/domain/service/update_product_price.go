package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type UpdateProductPriceParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	UserID    uuid.UUID
	Cost      *decimal.Decimal
	SalePrice *decimal.Decimal
	Reason    string
}

func (s *CatalogService) UpdateProductPrice(ctx context.Context, params UpdateProductPriceParams) (*repository.ProductListItem, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	if params.UserID == uuid.Nil {
		return nil, sharedErrors.ErrUnauthorized
	}
	if params.Cost == nil && params.SalePrice == nil {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "validation failed", 400).
			WithDetails(map[string]string{"prices": "debe enviar cost o sale_price"})
	}
	params.Reason = strings.TrimSpace(params.Reason)
	if utf8.RuneCountInString(params.Reason) > 160 {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "validation failed", 400).
			WithDetails(map[string]string{"reason": "admite como máximo 160 caracteres"})
	}

	var product *entity.Product
	err := s.products.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		product, err = s.findProduct(txCtx, params.CompanyID, params.ProductID)
		if err != nil {
			return err
		}
		previousCost, previousPrice := product.Cost, product.SalePrice
		previousUpdatedAt := product.UpdatedAt
		costPrice, salePrice := previousCost, previousPrice
		if params.Cost != nil {
			costPrice = *params.Cost
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
	return s.withCategoryName(ctx, product)
}
