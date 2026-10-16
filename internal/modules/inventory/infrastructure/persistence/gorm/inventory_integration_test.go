//go:build integration

package gorm_test

import (
	"context"
	"errors"
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

type stubProducts struct{ found bool }

func (s stubProducts) Exists(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return s.found, nil
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
}
