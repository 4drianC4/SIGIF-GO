package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/errors"
)

type ProductService struct {
	repo       repository.ProductRepository
	categoryRepo repository.CategoryRepository
	brandRepo    repository.BrandRepository
	unitRepo     repository.UnitRepository
	taxRepo      repository.TaxRepository
}

func NewProductService(
	repo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	brandRepo repository.BrandRepository,
	unitRepo repository.UnitRepository,
	taxRepo repository.TaxRepository,
) *ProductService {
	return &ProductService{
		repo:         repo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
		unitRepo:     unitRepo,
		taxRepo:      taxRepo,
	}
}

func (s *ProductService) Create(ctx context.Context, tenantID uuid.UUID, sku, name string, productType entity.ProductType) (*entity.Product, error) {
	exists, err := s.repo.ExistsBySKU(ctx, tenantID, sku)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(errors.CodeConflict, "product with this SKU already exists", 409)
	}

	product := entity.NewProduct(nil, tenantID, sku, name, productType)
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New(errors.CodeNotFound, "product not found", 404)
	}
	return product, nil
}

func (s *ProductService) GetBySKU(ctx context.Context, tenantID uuid.UUID, sku string) (*entity.Product, error) {
	product, err := s.repo.GetBySKU(ctx, tenantID, sku)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New(errors.CodeNotFound, "product not found", 404)
	}
	return product, nil
}

func (s *ProductService) List(ctx context.Context, tenantID uuid.UUID, offset, limit int, filters repository.ProductFilters) ([]*entity.Product, int64, error) {
	return s.repo.List(ctx, tenantID, offset, limit, filters)
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, sku, name, description string, productType entity.ProductType, categoryID, brandID, unitID, taxID *uuid.UUID, costPrice, salePrice, wholesalePrice, minPrice, maxPrice, minStock, maxStock, weight decimal.Decimal, dimensions, imageURL, images, attributes, tags string, settings entity.ProductSettings) (*entity.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New(errors.CodeNotFound, "product not found", 404)
	}

	if sku != product.SKU {
		exists, err := s.repo.ExistsBySKU(ctx, product.TenantID, sku)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New(errors.CodeConflict, "product with this SKU already exists", 409)
		}
	}

	product.Update(sku, name, description, productType, categoryID, brandID, unitID, taxID, costPrice, salePrice, wholesalePrice, minPrice, maxPrice, minStock, maxStock, weight, dimensions, imageURL, images, attributes, tags, settings, nil)
	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New(errors.CodeNotFound, "product not found", 404)
	}
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) Activate(ctx context.Context, id uuid.UUID) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New(errors.CodeNotFound, "product not found", 404)
	}
	product.Activate(nil)
	return s.repo.Update(ctx, product)
}

func (s *ProductService) Deactivate(ctx context.Context, id uuid.UUID) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New(errors.CodeNotFound, "product not found", 404)
	}
	product.Deactivate(nil)
	return s.repo.Update(ctx, product)
}

func (s *ProductService) AdjustStock(ctx context.Context, id uuid.UUID, quantity decimal.Decimal) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New(errors.CodeNotFound, "product not found", 404)
	}
	product.AdjustStock(quantity, nil)
	return s.repo.Update(ctx, product)
}

func (s *ProductService) SetStock(ctx context.Context, id uuid.UUID, quantity decimal.Decimal) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New(errors.CodeNotFound, "product not found", 404)
	}
	product.SetStock(quantity, nil)
	return s.repo.Update(ctx, product)
}

func (s *ProductService) GetLowStock(ctx context.Context, tenantID uuid.UUID) ([]*entity.Product, error) {
	return s.repo.GetLowStock(ctx, tenantID)
}