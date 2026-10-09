package testutil

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
)

var _ repository.CustomerRepository = (*MemoryCustomerRepository)(nil)

// MemoryCustomerRepository mirrors the filtering rules of the GORM repository.
type MemoryCustomerRepository struct {
	mu        sync.Mutex
	Customers map[uuid.UUID]*entity.Customer
	Writes    int
}

func NewMemoryCustomerRepository() *MemoryCustomerRepository {
	return &MemoryCustomerRepository{Customers: map[uuid.UUID]*entity.Customer{}}
}

func (r *MemoryCustomerRepository) Create(_ context.Context, customer *entity.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *customer
	r.Customers[customer.ID] = &copied
	r.Writes++
	return nil
}

func (r *MemoryCustomerRepository) Update(_ context.Context, customer *entity.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.Customers[customer.ID]
	if !ok || stored.CompanyID != customer.CompanyID || stored.IsDeleted() {
		return service.ErrCustomerNotFound
	}
	copied := *customer
	r.Customers[customer.ID] = &copied
	r.Writes++
	return nil
}

func (r *MemoryCustomerRepository) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// Get returns a copy of a stored customer, deleted ones included.
func (r *MemoryCustomerRepository) Get(id uuid.UUID) (entity.Customer, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.Customers[id]
	if !ok {
		return entity.Customer{}, false
	}
	return *c, true
}

// Count returns how many customers are stored, deleted ones included.
func (r *MemoryCustomerRepository) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Customers)
}

func (r *MemoryCustomerRepository) GetByID(_ context.Context, companyID, id uuid.UUID) (*entity.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	customer, ok := r.Customers[id]
	if !ok || customer.CompanyID != companyID || customer.IsDeleted() {
		return nil, nil
	}
	copied := *customer
	return &copied, nil
}

func (r *MemoryCustomerRepository) ExistsByDocument(_ context.Context, companyID uuid.UUID, docType entity.DocumentType, docNumber string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.Customers {
		if c.CompanyID == companyID && !c.IsDeleted() && c.DocumentType == docType &&
			c.DocumentNumber != nil && *c.DocumentNumber == docNumber {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryCustomerRepository) List(
	_ context.Context,
	companyID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]*entity.Customer, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	q := strings.ToLower(strings.TrimSpace(filter.Q))
	var matches []*entity.Customer
	for _, c := range r.Customers {
		if c.CompanyID != companyID || c.IsDeleted() {
			continue
		}
		if filter.Status != nil && c.Status != *filter.Status {
			continue
		}
		if q != "" && !containsAny(q, &c.LegalName, c.DocumentNumber, c.Phone, c.Email) {
			continue
		}
		copied := *c
		matches = append(matches, &copied)
	}

	sort.Slice(matches, func(i, j int) bool {
		a, b := matches[i], matches[j]
		if filter.SortOrder == repository.SortDesc {
			a, b = b, a
		}
		if filter.SortBy == repository.SortByLegalName {
			return a.LegalName < b.LegalName
		}
		return a.CreatedAt.Before(b.CreatedAt)
	})

	total := int64(len(matches))
	if offset >= len(matches) {
		return []*entity.Customer{}, total, nil
	}
	end := min(offset+limit, len(matches))
	return matches[offset:end], total, nil
}

func containsAny(q string, fields ...*string) bool {
	for _, f := range fields {
		if f != nil && strings.Contains(strings.ToLower(*f), q) {
			return true
		}
	}
	return false
}
