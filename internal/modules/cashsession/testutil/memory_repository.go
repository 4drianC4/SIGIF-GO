package testutil

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
)

var _ repository.SessionRepository = (*MemorySessionRepository)(nil)

// MemorySessionRepository mirrors the rules of the GORM repository, including
// the unique open indexes per cash register and per cashier.
type MemorySessionRepository struct {
	mu       sync.Mutex
	Sessions map[uuid.UUID]*entity.CashSession
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{Sessions: map[uuid.UUID]*entity.CashSession{}}
}

// Close marks a session closed, standing in for RF-17 in tests.
func (r *MemorySessionRepository) Close(id uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.Sessions[id]; ok {
		s.Status = entity.StatusClosed
	}
}

// OpenCount returns how many open sessions are stored.
func (r *MemorySessionRepository) OpenCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, s := range r.Sessions {
		if s.IsOpen() {
			count++
		}
	}
	return count
}

func (r *MemorySessionRepository) Create(_ context.Context, session *entity.CashSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.Sessions {
		if !s.IsOpen() {
			continue
		}
		if s.UserID == session.UserID {
			return service.ErrCashierHasOpenSession
		}
		if s.CashRegisterID == session.CashRegisterID {
			return service.ErrRegisterHasOpenSession
		}
	}
	copied := *session
	r.Sessions[session.ID] = &copied
	return nil
}

func (r *MemorySessionRepository) OpenByUser(_ context.Context, companyID, userID uuid.UUID) (*entity.CashSession, error) {
	return r.find(func(s *entity.CashSession) bool { return s.CompanyID == companyID && s.UserID == userID })
}

func (r *MemorySessionRepository) OpenByRegister(_ context.Context, companyID, cashRegisterID uuid.UUID) (*entity.CashSession, error) {
	return r.find(func(s *entity.CashSession) bool {
		return s.CompanyID == companyID && s.CashRegisterID == cashRegisterID
	})
}

func (r *MemorySessionRepository) find(match func(*entity.CashSession) bool) (*entity.CashSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.Sessions {
		if s.IsOpen() && match(s) {
			copied := *s
			return &copied, nil
		}
	}
	return nil, nil
}
