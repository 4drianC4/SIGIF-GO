package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
	"github.com/sigif/sigif-go/internal/modules/customer/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func TestEnsureAvailableForOperations(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewMockClock(time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC))
	repo := testutil.NewMemoryCustomerRepository()
	svc := service.NewCustomerService(repo, testutil.NewStubDependencyChecker(), clk)
	companyA, companyB := uuid.New(), uuid.New()

	seed := func(company uuid.UUID, status entity.CustomerStatus, deleted bool) uuid.UUID {
		c := entity.NewCustomer(clk, company, "Ferreteria Central", entity.DocumentTypeTaxID, nil, nil, nil)
		c.Status = status
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("create: %v", err)
		}
		if deleted {
			c.SoftDelete(clk)
			if err := repo.Update(ctx, c); err != nil {
				t.Fatalf("soft delete: %v", err)
			}
		}
		return c.ID
	}

	active := seed(companyA, entity.CustomerStatusActive, false)

	tests := []struct {
		name    string
		company uuid.UUID
		id      uuid.UUID
		want    error
	}{
		{"active customer of the company", companyA, active, nil},
		{"inactive customer", companyA, seed(companyA, entity.CustomerStatusInactive, false), service.ErrCustomerNotActive},
		{"blocked customer", companyA, seed(companyA, entity.CustomerStatusBlocked, false), service.ErrCustomerNotActive},
		{"deleted customer", companyA, seed(companyA, entity.CustomerStatusActive, true), service.ErrCustomerNotFound},
		{"customer of another company", companyA, seed(companyB, entity.CustomerStatusActive, false), service.ErrCustomerNotFound},
		{"unknown customer", companyA, uuid.New(), service.ErrCustomerNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := svc.EnsureAvailableForOperations(ctx, tt.company, tt.id)

			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if tt.want == nil && (customer == nil || customer.ID != tt.id) {
				t.Errorf("customer = %+v, want %s", customer, tt.id)
			}
			if tt.want != nil && customer != nil {
				t.Errorf("customer = %+v, want nil on error", customer)
			}
		})
	}
}

func TestCustomerIsActive(t *testing.T) {
	clk := clock.NewMockClock(time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC))
	customer := entity.NewCustomer(clk, uuid.New(), "Ferreteria Central", entity.DocumentTypeTaxID, nil, nil, nil)

	if !customer.IsActive() {
		t.Error("new customer must be active")
	}
	customer.Deactivate(clk)
	if customer.IsActive() {
		t.Error("inactive customer must not be active")
	}
	customer.Block(clk)
	if customer.IsActive() {
		t.Error("blocked customer must not be active")
	}
	customer.Activate(clk)
	customer.SoftDelete(clk)
	if customer.IsActive() {
		t.Error("deleted customer must not be active")
	}
}
