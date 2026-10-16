package testutil

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
)

var (
	UnitID       = uuid.MustParse("00000000-0000-4000-8000-000000000101")
	KilogramID   = uuid.MustParse("00000000-0000-4000-8000-000000000104")
	IVAGeneral   = "IVA general 13%"
	IVAGeneralID = uuid.MustParse("00000000-0000-4000-8000-000000000201")
)

type MemoryStore struct {
	mu         sync.Mutex
	Categories map[uuid.UUID]*entity.Category
	Products   map[uuid.UUID]*entity.Product
	Units      []*entity.UnitOfMeasure
	Taxes      []*entity.Tax

	PriceHistory *MemoryPriceHistoryRepository
	Stock        *MemoryStockRepository
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		PriceHistory: NewMemoryPriceHistoryRepository(),
		Stock:        &MemoryStockRepository{},
		Categories:   map[uuid.UUID]*entity.Category{},
		Products:     map[uuid.UUID]*entity.Product{},
		Units: []*entity.UnitOfMeasure{
			{ID: UnitID, Name: "Unit", Abbreviation: "UNIT"},
			{ID: KilogramID, Name: "Kilogram", Abbreviation: "KG"},
		},
		Taxes: []*entity.Tax{
			{ID: IVAGeneralID, Name: IVAGeneral, Percentage: decimal.NewFromInt(13)},
			{ID: uuid.MustParse("00000000-0000-4000-8000-000000000203"), Name: "Exempt", Percentage: decimal.Zero},
		},
	}
}

func (s *MemoryStore) CategoryRepository() repository.CategoryRepository  { return memoryCategories{s} }
func (s *MemoryStore) ProductRepository() repository.ProductRepository    { return memoryProducts{s} }
func (s *MemoryStore) UnitRepository() repository.UnitOfMeasureRepository { return memoryUnits{s} }
func (s *MemoryStore) TaxRepository() repository.TaxRepository            { return memoryTaxes{s} }

func (s *MemoryStore) PriceHistoryRepository() repository.PriceHistoryRepository {
	return s.PriceHistory
}

func (s *MemoryStore) StockRepository() repository.StockRepository { return s.Stock }

// ProductCount returns how many products are not deleted.
func (s *MemoryStore) ProductCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, p := range s.Products {
		if p.DeletedAt == nil {
			count++
		}
	}
	return count
}

type memoryCategories struct{ s *MemoryStore }

func (r memoryCategories) Create(_ context.Context, category *entity.Category) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	copied := *category
	r.s.Categories[category.ID] = &copied
	return nil
}

func (r memoryCategories) GetByID(_ context.Context, companyID, id uuid.UUID) (*entity.Category, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	category, ok := r.s.Categories[id]
	if !ok || category.CompanyID != companyID {
		return nil, nil
	}
	copied := *category
	return &copied, nil
}

func (r memoryCategories) ExistsByName(_ context.Context, companyID uuid.UUID, parentID *uuid.UUID, name string) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, c := range r.s.Categories {
		if c.CompanyID == companyID && sameParent(c.ParentID, parentID) && strings.EqualFold(c.Name, name) {
			return true, nil
		}
	}
	return false, nil
}

func (r memoryCategories) List(_ context.Context, companyID uuid.UUID) ([]repository.CategoryListItem, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var items []repository.CategoryListItem
	for _, c := range r.s.Categories {
		if c.CompanyID != companyID {
			continue
		}
		var count int64
		for _, p := range r.s.Products {
			if p.CategoryID == c.ID {
				count++
			}
		}
		var taxName *string
		for _, t := range r.s.Taxes {
			if c.DefaultTaxID != nil && t.ID == *c.DefaultTaxID {
				name := t.Name
				taxName = &name
			}
		}
		copied := *c
		items = append(items, repository.CategoryListItem{Category: &copied, ProductsCount: count, DefaultTaxName: taxName})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Category.Name < items[j].Category.Name })
	return items, nil
}

type memoryProducts struct{ s *MemoryStore }

func (r memoryProducts) Create(_ context.Context, product *entity.Product) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	copied := *product
	r.s.Products[product.ID] = &copied
	return nil
}

func (r memoryProducts) Update(_ context.Context, product *entity.Product) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	copied := *product
	r.s.Products[product.ID] = &copied
	return nil
}

// WithinTransaction restores the products as they were when fn fails, which is
// what rolling back the database transaction does.
func (r memoryProducts) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	r.s.mu.Lock()
	before := make(map[uuid.UUID]*entity.Product, len(r.s.Products))
	for id, product := range r.s.Products {
		copied := *product
		before[id] = &copied
	}
	r.s.mu.Unlock()
	if err := fn(ctx); err != nil {
		r.s.mu.Lock()
		r.s.Products = before
		r.s.mu.Unlock()
		return err
	}
	return nil
}

func (r memoryProducts) GetByID(_ context.Context, companyID, id uuid.UUID) (*entity.Product, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	product, ok := r.s.Products[id]
	if !ok || product.CompanyID != companyID || product.DeletedAt != nil {
		return nil, nil
	}
	copied := *product
	return &copied, nil
}

func (r memoryProducts) SetStatus(_ context.Context, companyID, id uuid.UUID, status entity.Status) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if product, ok := r.s.Products[id]; ok && product.CompanyID == companyID {
		product.Status = status
	}
	return nil
}

func (r memoryProducts) Delete(_ context.Context, companyID, id uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	// Soft delete, as the GORM repository does: the row keeps its SKU and
	// barcode reserved (HU-06-01).
	if product, ok := r.s.Products[id]; ok && product.CompanyID == companyID && product.DeletedAt == nil {
		now := time.Now().UTC()
		product.DeletedAt = &now
	}
	return nil
}

func (r memoryProducts) FindByName(_ context.Context, companyID uuid.UUID, name string, excludeID uuid.UUID) (*entity.Product, error) {
	return r.holder(companyID, excludeID, false, func(p *entity.Product) bool { return strings.EqualFold(p.Name, name) }), nil
}

func (r memoryProducts) FindBySKU(_ context.Context, companyID uuid.UUID, sku string, excludeID uuid.UUID) (*entity.Product, error) {
	return r.holder(companyID, excludeID, true, func(p *entity.Product) bool { return p.SKU == sku }), nil
}

func (r memoryProducts) FindByBarcode(_ context.Context, companyID uuid.UUID, barcode string, excludeID uuid.UUID) (*entity.Product, error) {
	return r.holder(companyID, excludeID, true, func(p *entity.Product) bool { return p.Barcode == barcode }), nil
}

// holder returns a copy of the product of the company that matches, skipping
// excludeID. Deleted products only count when withDeleted is set, and a live
// product is preferred over a deleted one.
func (r memoryProducts) holder(companyID, excludeID uuid.UUID, withDeleted bool, match func(*entity.Product) bool) *entity.Product {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var found *entity.Product
	for _, p := range r.s.Products {
		if p.CompanyID != companyID || p.ID == excludeID || !match(p) {
			continue
		}
		if p.DeletedAt != nil && (!withDeleted || found != nil) {
			continue
		}
		if found == nil || found.DeletedAt != nil {
			copied := *p
			found = &copied
		}
	}
	return found
}

func (r memoryProducts) List(_ context.Context, companyID uuid.UUID, filter repository.ProductFilter, offset, limit int) ([]repository.ProductListItem, int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	var items []repository.ProductListItem
	for _, p := range r.s.Products {
		if p.CompanyID != companyID || p.DeletedAt != nil {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(p.Name), search) &&
			!strings.Contains(strings.ToLower(p.SKU), search) && !strings.Contains(p.Barcode, search) {
			continue
		}
		if filter.CategoryID != nil && p.CategoryID != *filter.CategoryID {
			continue
		}
		if filter.Status != nil && p.Status != *filter.Status {
			continue
		}
		copied := *p
		items = append(items, repository.ProductListItem{Product: &copied, CategoryName: r.s.Categories[p.CategoryID].Name})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Product.Name < items[j].Product.Name })

	total := int64(len(items))
	if offset >= len(items) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], total, nil
}

func (r memoryProducts) Summary(_ context.Context, companyID uuid.UUID, noMovementSince time.Time) (repository.ProductSummary, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	summary := repository.ProductSummary{InventoryValue: decimal.Zero}
	for _, p := range r.s.Products {
		if p.CompanyID != companyID || p.DeletedAt != nil || p.Status != entity.StatusActive {
			continue
		}
		summary.ActiveProducts++
		summary.InventoryValue = summary.InventoryValue.Add(p.Stock.Mul(p.Cost))
		if p.Stock.IsPositive() && p.Stock.LessThanOrEqual(p.MinStock) {
			summary.LowStock++
		}
		if p.LastMovementAt.Before(noMovementSince) {
			summary.WithoutMovement++
		}
	}
	return summary, nil
}

type memoryUnits struct{ s *MemoryStore }

func (r memoryUnits) List(context.Context) ([]*entity.UnitOfMeasure, error) {
	return r.s.Units, nil
}

func (r memoryUnits) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	for _, u := range r.s.Units {
		if u.ID == id {
			return true, nil
		}
	}
	return false, nil
}

type memoryTaxes struct{ s *MemoryStore }

func (r memoryTaxes) List(context.Context) ([]*entity.Tax, error) {
	return r.s.Taxes, nil
}

func (r memoryTaxes) GetByName(_ context.Context, name string) (*entity.Tax, error) {
	for _, t := range r.s.Taxes {
		if strings.EqualFold(t.Name, name) {
			return t, nil
		}
	}
	return nil, nil
}

func sameParent(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

type MemoryPriceHistoryRepository struct {
	mu          sync.Mutex
	Entries     map[uuid.UUID][]*entity.PriceHistoryEntry
	RecordError error
}

func NewMemoryPriceHistoryRepository() *MemoryPriceHistoryRepository {
	return &MemoryPriceHistoryRepository{Entries: map[uuid.UUID][]*entity.PriceHistoryEntry{}}
}

func (r *MemoryPriceHistoryRepository) Record(_ context.Context, productID, userID uuid.UUID, previousCost, newCost, previousPrice, newPrice decimal.Decimal, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.RecordError != nil {
		return r.RecordError
	}
	r.Entries[productID] = append(r.Entries[productID], &entity.PriceHistoryEntry{
		ID: uuid.New(), UserID: userID, PreviousCost: previousCost, NewCost: newCost,
		PreviousPrice: previousPrice, NewPrice: newPrice, Reason: reason, CreatedAt: time.Now().UTC(),
	})
	return nil
}

func (r *MemoryPriceHistoryRepository) GetByProduct(_ context.Context, productID uuid.UUID, page, limit int) ([]*entity.PriceHistoryEntry, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entries := r.Entries[productID]
	result := make([]*entity.PriceHistoryEntry, 0)
	for i := (page - 1) * limit; i < len(entries) && i < page*limit; i++ {
		copied := *entries[len(entries)-1-i]
		result = append(result, &copied)
	}
	return result, int64(len(entries)), nil
}

type StockBalance struct {
	CompanyID  uuid.UUID
	ProductID  uuid.UUID
	BranchID   int64
	LocationID int64
	Quantity   decimal.Decimal
}

type MemoryStockRepository struct {
	Balances []StockBalance
	GetError error
}

func (r *MemoryStockRepository) GetTotalByProduct(_ context.Context, companyID, productID uuid.UUID) (decimal.Decimal, error) {
	if r.GetError != nil {
		return decimal.Zero, r.GetError
	}
	total := decimal.Zero
	for _, balance := range r.Balances {
		if balance.CompanyID == companyID && balance.ProductID == productID {
			total = total.Add(balance.Quantity)
		}
	}
	return total, nil
}
