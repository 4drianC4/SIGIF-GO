package service

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

// CatalogService expone los casos de uso de dominio del catálogo (productos y categorías).
// Cada método está en su propio archivo (create_product.go, create_category.go).
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
