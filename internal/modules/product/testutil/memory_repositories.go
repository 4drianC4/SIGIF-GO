// Package testutil contiene repositorios en memoria para probar el módulo product sin base de datos.
package testutil

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type MemoryCategoryRepository struct {
	mu         sync.Mutex
	Categories map[uuid.UUID]*entity.Category
}

func NewMemoryCategoryRepository() *MemoryCategoryRepository {
	return &MemoryCategoryRepository{Categories: map[uuid.UUID]*entity.Category{}}
}

func (r *MemoryCategoryRepository) Create(_ context.Context, category *entity.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *category
	r.Categories[category.ID] = &copied
	return nil
}

func (r *MemoryCategoryRepository) GetByID(_ context.Context, tenantID, id uuid.UUID) (*entity.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	category, ok := r.Categories[id]
	if !ok || category.TenantID != tenantID {
		return nil, nil
	}
	copied := *category
	return &copied, nil
}

func (r *MemoryCategoryRepository) ExistsByName(_ context.Context, tenantID uuid.UUID, name string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, category := range r.Categories {
		if category.TenantID == tenantID && strings.EqualFold(category.Name, name) {
			return true, nil
		}
	}
	return false, nil
}

type MemoryProductRepository struct {
	mu       sync.Mutex
	Products map[uuid.UUID]*entity.Product
}

func NewMemoryProductRepository() *MemoryProductRepository {
	return &MemoryProductRepository{Products: map[uuid.UUID]*entity.Product{}}
}

func (r *MemoryProductRepository) Create(_ context.Context, product *entity.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *product
	r.Products[product.ID] = &copied
	return nil
}

func (r *MemoryProductRepository) ExistsBySKU(_ context.Context, tenantID uuid.UUID, sku string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, product := range r.Products {
		if product.TenantID == tenantID && product.SKU == sku {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryProductRepository) ExistsByBarcode(_ context.Context, tenantID uuid.UUID, barcode string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, product := range r.Products {
		if product.TenantID == tenantID && product.Barcode == barcode {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryProductRepository) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Products)
}
