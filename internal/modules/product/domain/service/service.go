package service

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type CatalogService struct {
	products   repository.ProductRepository
	categories repository.CategoryRepository
	units      repository.UnitOfMeasureRepository
	taxes      repository.TaxRepository
	clock      clock.Clock
}

func NewCatalogService(
	products repository.ProductRepository,
	categories repository.CategoryRepository,
	units repository.UnitOfMeasureRepository,
	taxes repository.TaxRepository,
	clock clock.Clock,
) *CatalogService {
	return &CatalogService{
		products:   products,
		categories: categories,
		units:      units,
		taxes:      taxes,
		clock:      clock,
	}
}
