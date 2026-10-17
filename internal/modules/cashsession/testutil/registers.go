package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	cashRegisterEntity "github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	cashRegisterService "github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	cashRegisterTestutil "github.com/sigif/sigif-go/internal/modules/cashregister/testutil"
	"github.com/sigif/sigif-go/internal/modules/cashsession/application/adapter"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

// Registers runs the real cash register module (RF-10) on its in-memory
// repository, so cash session tests exercise the same assignment rules.
type Registers struct {
	Repo    *cashRegisterTestutil.MemoryCashRegisterRepository
	Service *cashRegisterService.CashRegisterService
	Access  repository.RegisterAccess
}

func NewRegisters() *Registers {
	repo := cashRegisterTestutil.NewMemoryCashRegisterRepository()
	clk := clock.NewMockClock(time.Date(2026, 10, 16, 8, 0, 0, 0, time.UTC))
	svc := cashRegisterService.NewCashRegisterService(repo, cashRegisterTestutil.NewStubCashierDirectory(), clk)
	return &Registers{Repo: repo, Service: svc, Access: adapter.NewRegisterAccessAdapter(svc)}
}

// Add creates a cash register of the company.
func (r *Registers) Add(companyID uuid.UUID, code string, active bool) *cashRegisterEntity.CashRegister {
	status := cashRegisterEntity.CashRegisterStatusActive
	if !active {
		status = cashRegisterEntity.CashRegisterStatusInactive
	}
	return r.Repo.AddRegister(companyID, code, status)
}

// Assign gives the cash register to the cashier and returns the assignment id.
func (r *Registers) Assign(t *testing.T, companyID, userID, cashRegisterID uuid.UUID) uuid.UUID {
	t.Helper()
	view, err := r.Service.Assign(context.Background(), companyID, userID, cashRegisterID)
	if err != nil {
		t.Fatalf("assign cash register: %v", err)
	}
	return view.Assignment.ID
}

// Release frees an assignment as the cashier who holds it.
func (r *Registers) Release(t *testing.T, companyID, userID, assignmentID uuid.UUID) {
	t.Helper()
	_, err := r.Service.Release(context.Background(), cashRegisterService.ReleaseParams{
		CompanyID: companyID, ActorID: userID, AssignmentID: assignmentID,
	})
	if err != nil {
		t.Fatalf("release cash register: %v", err)
	}
}
