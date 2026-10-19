//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/service"
	inventoryGorm "github.com/sigif/sigif-go/internal/modules/inventory/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/database/dbtest"
	"github.com/sigif/sigif-go/internal/shared/database/migrations"
)

type stubProducts struct {
	found bool
	stock decimal.Decimal
}

func (s stubProducts) Stock(context.Context, uuid.UUID, uuid.UUID) (decimal.Decimal, bool, error) {
	return s.stock, s.found, nil
}

func openInventory(t *testing.T) *sharedDatabase.Database {
	t.Helper()
	dsn := dbtest.NewSchemaDSN(t)
	if err := migrations.Up(dsn); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	db, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatalf("database pool: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	return &sharedDatabase.Database{DB: db}
}

func TestRegisterMovementUpdatesBalanceAndProduct(t *testing.T) {
	db := openInventory(t)
	ctx := context.Background()
	companyID := uuid.New()

	categoryID := uuid.New()
	if err := db.DB.Exec(`INSERT INTO categories (id, company_id, name, created_at) VALUES (?, ?, 'Bebidas', now())`, categoryID, companyID).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	var unitRaw string
	db.DB.Raw(`SELECT id FROM units_of_measure LIMIT 1`).Scan(&unitRaw)
	unitID := uuid.MustParse(unitRaw)
	productID := uuid.New()
	oldMovement := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := db.DB.Exec(`INSERT INTO products (id, company_id, category_id, unit_of_measure_id, name, cost, sale_price, stock, min_stock, status, last_movement_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'Producto', 1, 2, 0, 0, 'active', ?, now(), now())`, productID, companyID, categoryID, unitID, oldMovement).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}

	branches := inventoryGorm.NewBranchLocationGormRepository(db)
	balances := inventoryGorm.NewStockBalanceGormRepository(db)
	movements := inventoryGorm.NewMovementGormRepository(db)
	svc := service.NewInventoryService(branches, balances, movements, stubProducts{found: true}, clock.NewRealClock())

	result, err := svc.RegisterMovement(ctx, service.RegisterMovementParams{
		CompanyID: companyID, ProductID: productID, Type: entity.MovementIn, Quantity: decimal.NewFromInt(5),
	})
	if err != nil {
		t.Fatalf("register in: %v", err)
	}
	if !result.Balance.Equal(decimal.NewFromInt(5)) || result.Movement.BranchID == 0 || result.Movement.LocationID == 0 {
		t.Fatalf("result = %+v", result)
	}

	items, err := balances.TotalsByBranch(ctx, companyID, productID)
	if err != nil || len(items) != 1 || !items[0].Quantity.Equal(decimal.NewFromInt(5)) {
		t.Fatalf("balances = %+v, %v", items, err)
	}
	var lastMovement time.Time
	db.DB.Raw(`SELECT last_movement_at FROM products WHERE id = ?`, productID).Scan(&lastMovement)
	if !lastMovement.After(oldMovement) {
		t.Errorf("last_movement_at = %v, want after %v", lastMovement, oldMovement)
	}

	list, total, err := movements.ListByProduct(ctx, companyID, productID, 0, 20)
	if err != nil || total != 1 || len(list) != 1 || list[0].Type != entity.MovementIn {
		t.Fatalf("movements = %+v total=%d err=%v", list, total, err)
	}

	// Going below zero is rejected and rolls back the second movement.
	_, err = svc.RegisterMovement(ctx, service.RegisterMovementParams{
		CompanyID: companyID, ProductID: productID, BranchID: result.Movement.BranchID, LocationID: result.Movement.LocationID,
		Type: entity.MovementOut, Quantity: decimal.NewFromInt(99),
	})
	if !errors.Is(err, service.ErrInsufficientStock) {
		t.Fatalf("out beyond stock error = %v, want ErrInsufficientStock", err)
	}
	if _, total, _ := movements.ListByProduct(ctx, companyID, productID, 0, 20); total != 1 {
		t.Errorf("rejected movement must not be recorded, total = %d", total)
	}

	// Later movements add to the same balance row instead of creating another one.
	same := service.RegisterMovementParams{CompanyID: companyID, ProductID: productID, BranchID: result.Movement.BranchID, LocationID: result.Movement.LocationID}
	same.Type, same.Quantity = entity.MovementIn, decimal.NewFromInt(3)
	more, err := svc.RegisterMovement(ctx, same)
	if err != nil || !more.Balance.Equal(decimal.NewFromInt(8)) {
		t.Fatalf("second entry: result = %+v, error = %v; want balance 8", more, err)
	}
	same.Type, same.Quantity = entity.MovementOut, decimal.NewFromInt(2)
	less, err := svc.RegisterMovement(ctx, same)
	if err != nil || !less.Balance.Equal(decimal.NewFromInt(6)) {
		t.Fatalf("exit within stock: result = %+v, error = %v; want balance 6", less, err)
	}
	if items, err := balances.TotalsByBranch(ctx, companyID, productID); err != nil || len(items) != 1 || !items[0].Quantity.Equal(decimal.NewFromInt(6)) {
		t.Errorf("balances = %+v, %v; want one row with 6", items, err)
	}

	// The whole balance can be taken out, and not a unit more.
	same.Quantity = decimal.NewFromInt(6)
	if empty, err := svc.RegisterMovement(ctx, same); err != nil || !empty.Balance.IsZero() {
		t.Fatalf("exit of the whole balance: result = %+v, error = %v; want balance 0", empty, err)
	}
	same.Quantity = decimal.NewFromInt(1)
	if _, err := svc.RegisterMovement(ctx, same); !errors.Is(err, service.ErrInsufficientStock) {
		t.Errorf("exit from an empty balance: error = %v, want ErrInsufficientStock", err)
	}

	// A company cannot move the stock of a product that belongs to another one.
	if _, err := balances.ApplyDelta(ctx, uuid.New(), productID, result.Movement.BranchID, result.Movement.LocationID, decimal.Zero, decimal.NewFromInt(1)); !errors.Is(err, service.ErrProductNotFound) {
		t.Errorf("product of another company: error = %v, want ErrProductNotFound", err)
	}
}

// Exits registered at the same time must never take out more than there is.
func TestConcurrentExitsNeverOversell(t *testing.T) {
	db := openInventory(t)
	ctx := context.Background()
	companyID := uuid.New()

	categoryID := uuid.New()
	if err := db.DB.Exec(`INSERT INTO categories (id, company_id, name, created_at) VALUES (?, ?, 'Bebidas', now())`, categoryID, companyID).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	var unitRaw string
	db.DB.Raw(`SELECT id FROM units_of_measure LIMIT 1`).Scan(&unitRaw)
	productID := uuid.New()
	if err := db.DB.Exec(`INSERT INTO products (id, company_id, category_id, unit_of_measure_id, name, cost, sale_price, stock, min_stock, status, last_movement_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'Producto', 1, 2, 0, 0, 'active', now(), now(), now())`, productID, companyID, categoryID, uuid.MustParse(unitRaw)).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}
	balances := inventoryGorm.NewStockBalanceGormRepository(db)
	svc := service.NewInventoryService(inventoryGorm.NewBranchLocationGormRepository(db), balances,
		inventoryGorm.NewMovementGormRepository(db), stubProducts{found: true}, clock.NewRealClock())

	stocked, err := svc.RegisterMovement(ctx, service.RegisterMovementParams{
		CompanyID: companyID, ProductID: productID, Type: entity.MovementIn, Quantity: decimal.NewFromInt(5),
	})
	if err != nil {
		t.Fatal(err)
	}

	const attempts = 12
	var wg sync.WaitGroup
	var mu sync.Mutex
	succeeded, refused := 0, 0
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.RegisterMovement(ctx, service.RegisterMovementParams{
				CompanyID: companyID, ProductID: productID, BranchID: stocked.Movement.BranchID, LocationID: stocked.Movement.LocationID,
				Type: entity.MovementOut, Quantity: decimal.NewFromInt(1),
			})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				succeeded++
			case errors.Is(err, service.ErrInsufficientStock):
				refused++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	items, err := balances.TotalsByBranch(ctx, companyID, productID)
	if err != nil || len(items) != 1 || !items[0].Quantity.IsZero() {
		t.Errorf("balances = %+v, %v; want one row with 0", items, err)
	}
	if succeeded != 5 || refused != attempts-5 {
		t.Errorf("%d exits succeeded and %d were refused; want 5 and %d", succeeded, refused, attempts-5)
	}
}
