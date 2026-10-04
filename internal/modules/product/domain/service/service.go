package service

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type CatalogService struct {
	products   repository.ProductRepository
	categories repository.CategoryRepository
	clock      clock.Clock
}

func NewCatalogService(products repository.ProductRepository, categories repository.CategoryRepository, clock clock.Clock) *CatalogService {
	return &CatalogService{
		products:   products,
		categories: categories,
		clock:      clock,
	}
}
