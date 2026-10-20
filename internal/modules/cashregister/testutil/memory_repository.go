package testutil

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
)

var _ repository.CashRegisterRepository = (*MemoryCashRegisterRepository)(nil)

// MemoryCashRegisterRepository mirrors the rules of the GORM repository,
// including the unique active indexes.
type MemoryCashRegisterRepository struct {
	mu          sync.Mutex
	Registers   map[uuid.UUID]*entity.CashRegister
	Assignments map[uuid.UUID]*entity.Assignment
	Writes      int
}

func NewMemoryCashRegisterRepository() *MemoryCashRegisterRepository {
	return &MemoryCashRegisterRepository{
		Registers:   map[uuid.UUID]*entity.CashRegister{},
		Assignments: map[uuid.UUID]*entity.Assignment{},
	}
}

// AddRegister stores a cash register directly, as the seed does.
func (r *MemoryCashRegisterRepository) AddRegister(companyID uuid.UUID, code string, status entity.CashRegisterStatus) *entity.CashRegister {
	r.mu.Lock()
	defer r.mu.Unlock()
	register := &entity.CashRegister{
		ID:        uuid.New(),
		CompanyID: companyID,
		Code:      code,
		Name:      "Caja " + code,
		Status:    status,
		CreatedAt: time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC),
	}
	r.Registers[register.ID] = register
	copied := *register
	return &copied
}

// Assignment returns a copy of a stored assignment.
func (r *MemoryCashRegisterRepository) Assignment(id uuid.UUID) (entity.Assignment, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.Assignments[id]
	if !ok {
		return entity.Assignment{}, false
	}
	return *a, true
}

// ActiveCount returns how many active assignments are stored.
func (r *MemoryCashRegisterRepository) ActiveCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, a := range r.Assignments {
		if a.IsActive() {
			count++
		}
	}
	return count
}

func (r *MemoryCashRegisterRepository) ListRegisters(
	_ context.Context,
	companyID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]repository.RegisterWithAssignment, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var matched []repository.RegisterWithAssignment
	for _, register := range r.Registers {
		if register.CompanyID != companyID || !register.IsActive() {
			continue
		}
		assignment := r.activeByRegister(register.ID)
		if filter.Availability != nil {
			inUse := assignment != nil
			if (*filter.Availability == entity.AvailabilityInUse) != inUse {
				continue
			}
		}
		copied := *register
		matched = append(matched, repository.RegisterWithAssignment{Register: &copied, Assignment: assignment})
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].Register.Code < matched[j].Register.Code })

	total := int64(len(matched))
	if offset >= len(matched) {
		return []repository.RegisterWithAssignment{}, total, nil
	}
	end := min(offset+limit, len(matched))
	return matched[offset:end], total, nil
}

func (r *MemoryCashRegisterRepository) GetRegister(_ context.Context, companyID, id uuid.UUID) (*entity.CashRegister, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	register, ok := r.Registers[id]
	if !ok || register.CompanyID != companyID {
		return nil, nil
	}
	copied := *register
	return &copied, nil
}

func (r *MemoryCashRegisterRepository) GetAssignment(_ context.Context, companyID, id uuid.UUID) (*entity.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.Assignments[id]
	if !ok || a.CompanyID != companyID {
		return nil, nil
	}
	copied := *a
	return &copied, nil
}

func (r *MemoryCashRegisterRepository) ActiveAssignmentByRegister(_ context.Context, companyID, cashRegisterID uuid.UUID) (*entity.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a := r.activeByRegister(cashRegisterID)
	if a == nil || a.CompanyID != companyID {
		return nil, nil
	}
	return a, nil
}

func (r *MemoryCashRegisterRepository) ActiveAssignmentByUser(_ context.Context, companyID, userID uuid.UUID) (*entity.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a := r.activeByUser(userID)
	if a == nil || a.CompanyID != companyID {
		return nil, nil
	}
	return a, nil
}

func (r *MemoryCashRegisterRepository) CreateAssignment(_ context.Context, assignment *entity.Assignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeByUser(assignment.UserID) != nil {
		return service.ErrCashierAlreadyAssigned
	}
	if r.activeByRegister(assignment.CashRegisterID) != nil {
		return service.ErrCashRegisterInUse
	}
	copied := *assignment
	r.Assignments[assignment.ID] = &copied
	r.Writes++
	return nil
}

func (r *MemoryCashRegisterRepository) ReleaseAssignment(_ context.Context, assignment *entity.Assignment) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.Assignments[assignment.ID]
	if !ok || stored.CompanyID != assignment.CompanyID || !stored.IsActive() {
		return false, nil
	}
	copied := *assignment
	r.Assignments[assignment.ID] = &copied
	r.Writes++
	return true, nil
}

func (r *MemoryCashRegisterRepository) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (r *MemoryCashRegisterRepository) activeByRegister(cashRegisterID uuid.UUID) *entity.Assignment {
	for _, a := range r.Assignments {
		if a.CashRegisterID == cashRegisterID && a.IsActive() {
			copied := *a
			return &copied
		}
	}
	return nil
}

func (r *MemoryCashRegisterRepository) activeByUser(userID uuid.UUID) *entity.Assignment {
	for _, a := range r.Assignments {
		if a.UserID == userID && a.IsActive() {
			copied := *a
			return &copied
		}
	}
	return nil
}
