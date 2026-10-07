package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
)

type GetProductParams struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
}

type ListProductsParams struct {
	CompanyID  uuid.UUID
	Name       string
	CategoryID *uuid.UUID
	Status     *entity.Status
	Page       int
	Limit      int
}

func (s *CatalogService) GetProduct(ctx context.Context, params GetProductParams) (*entity.Product, error) {
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
	return product, nil
}

func (s *CatalogService) ListProducts(ctx context.Context, params ListProductsParams) ([]*entity.Product, int64, error) {
	if params.CompanyID == uuid.Nil {
		return nil, 0, ErrCompanyRequired
	}
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	return s.products.GetAll(ctx, params.CompanyID, repository.ProductFilters{
		Name:       params.Name,
		CategoryID: params.CategoryID,
		Status:     params.Status,
	}, params.Page, params.Limit)
}
