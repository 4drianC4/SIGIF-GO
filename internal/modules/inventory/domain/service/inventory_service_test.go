package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/service"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	companyID = uuid.New()
	productID = uuid.New()
)

type fakeProducts struct {
	found bool
	err   error
}

func (f fakeProducts) Exists(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return f.found, f.err
}

type fakeBranches struct {
	defaultCalls int
	branches     []*entity.Branch
}

func (f *fakeBranches) ListBranches(context.Context, uuid.UUID) ([]*entity.Branch, error) {
	return f.branches, nil
}

func (f *fakeBranches) EnsureDefault(context.Context, uuid.UUID) (int64, int64, error) {
	f.defaultCalls++
	return 1, 1, nil
}

type fakeBalances struct {
	quantity     decimal.Decimal
	touchErr     error
	touched      *time.Time
	insufficient bool
}

func (f *fakeBalances) ApplyDelta(_ context.Context, _, _ uuid.UUID, _, _ int64, delta decimal.Decimal) (decimal.Decimal, error) {
	if f.insufficient {
		return decimal.Zero, service.ErrInsufficientStock
	}
	f.quantity = f.quantity.Add(delta)
	return f.quantity, nil
}

func (f *fakeBalances) TotalsByBranch(context.Context, uuid.UUID, uuid.UUID) ([]repository.StockBalanceItem, error) {
	return nil, nil
}

func (f *fakeBalances) TouchLastMovement(_ context.Context, _, _ uuid.UUID, at time.Time) error {
	if f.touchErr != nil {
		return f.touchErr
	}
	f.touched = &at
	return nil
}

func (f *fakeBalances) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeMovements struct {
	created []*entity.Movement
	err     error
}

func (f *fakeMovements) Create(_ context.Context, m *entity.Movement) error {
	if f.err != nil {
		return f.err
	}
	f.created = append(f.created, m)
	return nil
}

func (f *fakeMovements) ListByProduct(context.Context, uuid.UUID, uuid.UUID, int, int) ([]*entity.Movement, int64, error) {
	return f.created, int64(len(f.created)), nil
}

func newService(products repository.ProductReader, balances *fakeBalances, movements *fakeMovements) (*service.InventoryService, *fakeBranches) {
	branches := &fakeBranches{}
	return service.NewInventoryService(branches, balances, movements, products, clock.NewMockClock(time.Now().UTC())), branches
}

func TestRegisterMovementAppliesDeltaAndRecords(t *testing.T) {
	balances := &fakeBalances{}
	movements := &fakeMovements{}
	svc, branches := newService(fakeProducts{found: true}, balances, movements)

	result, err := svc.RegisterMovement(context.Background(), service.RegisterMovementParams{
		CompanyID: companyID, ProductID: productID, Type: entity.MovementIn, Quantity: decimal.NewFromInt(5),
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if !result.Balance.Equal(decimal.NewFromInt(5)) || len(movements.created) != 1 {
		t.Fatalf("result = %+v, movements = %d", result, len(movements.created))
	}
	if branches.defaultCalls != 1 {
		t.Errorf("default branch calls = %d, want 1", branches.defaultCalls)
	}
	if balances.touched == nil {
		t.Errorf("last movement was not touched")
	}
}

func TestRegisterMovementOutDecreasesStock(t *testing.T) {
	balances := &fakeBalances{quantity: decimal.NewFromInt(10)}
	svc, _ := newService(fakeProducts{found: true}, balances, &fakeMovements{})

	result, err := svc.RegisterMovement(context.Background(), service.RegisterMovementParams{
		CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 1,
		Type: entity.MovementOut, Quantity: decimal.NewFromInt(4),
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if !result.Balance.Equal(decimal.NewFromInt(6)) {
		t.Errorf("balance = %s, want 6", result.Balance)
	}
	if !result.Movement.Quantity.Equal(decimal.NewFromInt(-4)) {
		t.Errorf("stored quantity = %s, want -4", result.Movement.Quantity)
	}
}

func TestRegisterMovementRejectsUnknownProduct(t *testing.T) {
	svc, _ := newService(fakeProducts{found: false}, &fakeBalances{}, &fakeMovements{})
	_, err := svc.RegisterMovement(context.Background(), service.RegisterMovementParams{
		CompanyID: companyID, ProductID: productID, Type: entity.MovementIn, Quantity: decimal.NewFromInt(1),
	})
	if !errors.Is(err, service.ErrProductNotFound) {
		t.Fatalf("error = %v, want ErrProductNotFound", err)
	}
}

func TestRegisterMovementValidatesTypeAsBadRequest(t *testing.T) {
	svc, _ := newService(fakeProducts{found: true}, &fakeBalances{}, &fakeMovements{})
	_, err := svc.RegisterMovement(context.Background(), service.RegisterMovementParams{
		CompanyID: companyID, ProductID: productID, Type: entity.MovementType("bogus"), Quantity: decimal.NewFromInt(1),
	})
	if !sharedErrors.Is(err, sharedErrors.CodeValidation) {
		t.Fatalf("error = %v, want validation", err)
	}
}

func TestRegisterMovementPropagatesInsufficientStock(t *testing.T) {
	balances := &fakeBalances{insufficient: true}
	movements := &fakeMovements{}
	svc, _ := newService(fakeProducts{found: true}, balances, movements)

	_, err := svc.RegisterMovement(context.Background(), service.RegisterMovementParams{
		CompanyID: companyID, ProductID: productID, Type: entity.MovementOut, Quantity: decimal.NewFromInt(1),
	})
	if !errors.Is(err, service.ErrInsufficientStock) {
		t.Fatalf("error = %v, want ErrInsufficientStock", err)
	}
	if len(movements.created) != 0 {
		t.Errorf("movement must not be recorded when the delta fails")
	}
}
