package service

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type CatalogService struct {
	products     repository.ProductRepository
	categories   repository.CategoryRepository
	priceHistory repository.PriceHistoryRepository
	stock        repository.StockRepository
	clock        clock.Clock
}

func NewCatalogService(
	products repository.ProductRepository,
	categories repository.CategoryRepository,
	priceHistory repository.PriceHistoryRepository,
	stock repository.StockRepository,
	clock clock.Clock,
) *CatalogService {
	return &CatalogService{
		products:     products,
		categories:   categories,
		priceHistory: priceHistory,
		stock:        stock,
		clock:        clock,
	}
}
