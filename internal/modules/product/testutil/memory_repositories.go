package testutil

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
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

func (r *MemoryCategoryRepository) GetByID(_ context.Context, companyID, id uuid.UUID) (*entity.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	category, ok := r.Categories[id]
	if !ok || category.CompanyID != companyID {
		return nil, nil
	}
	copied := *category
	return &copied, nil
}

func (r *MemoryCategoryRepository) ExistsByName(_ context.Context, companyID uuid.UUID, name string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, category := range r.Categories {
		if category.CompanyID == companyID && strings.EqualFold(category.Name, name) {
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

func (r *MemoryProductRepository) Update(_ context.Context, product *entity.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.Products[product.ID]; !ok {
		return service.ErrProductNotFound
	}
	copied := *product
	r.Products[product.ID] = &copied
	return nil
}

func (r *MemoryProductRepository) GetByID(_ context.Context, companyID, id uuid.UUID) (*entity.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	product, ok := r.Products[id]
	if !ok || product.CompanyID != companyID {
		return nil, nil
	}
	copied := *product
	return &copied, nil
}

func (r *MemoryProductRepository) GetAll(_ context.Context, companyID uuid.UUID, filters repository.ProductFilters, page, limit int) ([]*entity.Product, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var matched []*entity.Product
	for _, p := range r.Products {
		if p.CompanyID != companyID {
			continue
		}
		if filters.Name != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(filters.Name)) {
			continue
		}
		if filters.CategoryID != nil && p.CategoryID != *filters.CategoryID {
			continue
		}
		if filters.Status != nil && p.Status != *filters.Status {
			continue
		}
		copied := *p
		matched = append(matched, &copied)
	}

	total := int64(len(matched))
	offset := (page - 1) * limit
	if offset >= len(matched) {
		return []*entity.Product{}, total, nil
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}
	return matched[offset:end], total, nil
}

func (r *MemoryProductRepository) SetStatus(_ context.Context, companyID, id uuid.UUID, status entity.Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	product, ok := r.Products[id]
	if !ok || product.CompanyID != companyID {
		return service.ErrProductNotFound
	}
	product.Status = status
	return nil
}

func (r *MemoryProductRepository) Delete(_ context.Context, companyID, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	product, ok := r.Products[id]
	if !ok || product.CompanyID != companyID {
		return service.ErrProductNotFound
	}
	delete(r.Products, id)
	return nil
}

func (r *MemoryProductRepository) ExistsBySKU(_ context.Context, companyID uuid.UUID, sku string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, product := range r.Products {
		if product.CompanyID == companyID && product.SKU == sku {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryProductRepository) ExistsBySKUExcluding(_ context.Context, companyID uuid.UUID, sku string, excludeID uuid.UUID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, product := range r.Products {
		if product.CompanyID == companyID && product.SKU == sku && product.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryProductRepository) ExistsByBarcode(_ context.Context, companyID uuid.UUID, barcode string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, product := range r.Products {
		if product.CompanyID == companyID && product.Barcode == barcode {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryProductRepository) ExistsByBarcodeExcluding(_ context.Context, companyID uuid.UUID, barcode string, excludeID uuid.UUID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, product := range r.Products {
		if product.CompanyID == companyID && product.Barcode == barcode && product.ID != excludeID {
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
