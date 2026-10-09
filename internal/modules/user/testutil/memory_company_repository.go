package testutil

import (
	"context"
	"sync"

	"github.com/google/uuid"

	companyEntity "github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	companyRepository "github.com/sigif/sigif-go/internal/modules/company/domain/repository"
)

var _ companyRepository.CompanyRepository = (*MemoryCompanyRepository)(nil)

// MemoryCompanyRepository is a minimal in-memory company repo for tests that
// only need the CompanyRepository port (e.g. wiring the user service).
type MemoryCompanyRepository struct {
	mu  sync.Mutex
	ids map[uuid.UUID]bool
}

func NewMemoryCompanyRepository() *MemoryCompanyRepository {
	return &MemoryCompanyRepository{ids: map[uuid.UUID]bool{}}
}

func (r *MemoryCompanyRepository) ExistsByID(_ context.Context, id uuid.UUID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ids[id], nil
}

func (r *MemoryCompanyRepository) List(context.Context) ([]companyEntity.Company, error) {
	return nil, nil
}
