package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	"github.com/sigif/sigif-go/internal/modules/cashregister/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func newService() (*service.CashRegisterService, *testutil.MemoryCashRegisterRepository) {
	repo := testutil.NewMemoryCashRegisterRepository()
	clk := clock.NewMockClock(time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC))
	return service.NewCashRegisterService(repo, testutil.NewStubCashierDirectory(), clk), repo
}

func TestEnsureCashierOwnsRegister(t *testing.T) {
	ctx := context.Background()
	svc, repo := newService()
	companyA, companyB := uuid.New(), uuid.New()
	holder, other := uuid.New(), uuid.New()

	held := repo.AddRegister(companyA, "CAJA-01", entity.CashRegisterStatusActive)
	free := repo.AddRegister(companyA, "CAJA-02", entity.CashRegisterStatusActive)
	released := repo.AddRegister(companyA, "CAJA-03", entity.CashRegisterStatusActive)
	foreign := repo.AddRegister(companyB, "CAJA-B1", entity.CashRegisterStatusActive)

	if _, err := svc.Assign(ctx, companyA, holder, held.ID); err != nil {
		t.Fatalf("assign held: %v", err)
	}
	view, err := svc.Assign(ctx, companyA, other, released.ID)
	if err != nil {
		t.Fatalf("assign released: %v", err)
	}
	if _, err := svc.Release(ctx, service.ReleaseParams{CompanyID: companyA, ActorID: other, AssignmentID: view.Assignment.ID}); err != nil {
		t.Fatalf("release: %v", err)
	}

	tests := []struct {
		name     string
		company  uuid.UUID
		user     uuid.UUID
		register uuid.UUID
		want     error
	}{
		{"the assigned cashier passes", companyA, holder, held.ID, nil},
		{"another cashier is rejected", companyA, other, held.ID, service.ErrCashRegisterInUse},
		{"a free register is not owned", companyA, holder, free.ID, service.ErrCashRegisterNotOwned},
		{"a released register is not owned", companyA, other, released.ID, service.ErrCashRegisterNotOwned},
		{"register of another company", companyA, holder, foreign.ID, service.ErrCashRegisterNotFound},
		{"same register checked from another company", companyB, holder, held.ID, service.ErrCashRegisterNotFound},
		{"unknown register", companyA, holder, uuid.New(), service.ErrCashRegisterNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := svc.EnsureCashierOwnsRegister(ctx, tt.company, tt.user, tt.register); !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestAssignmentRelease(t *testing.T) {
	clk := clock.NewMockClock(time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC))
	cashier, supervisor := uuid.New(), uuid.New()
	a := entity.NewAssignment(clk, uuid.New(), uuid.New(), cashier)

	if !a.IsActive() || !a.BelongsTo(cashier) || a.BelongsTo(supervisor) || a.ReleasedAt != nil {
		t.Fatalf("new assignment = %+v, want active and owned by the cashier", a)
	}
	a.Release(clk, supervisor)
	if a.IsActive() || a.Status != entity.AssignmentStatusReleased || a.ReleasedAt == nil ||
		a.ReleasedBy == nil || *a.ReleasedBy != supervisor || a.UpdatedAt == nil {
		t.Errorf("released assignment = %+v", a)
	}
}
