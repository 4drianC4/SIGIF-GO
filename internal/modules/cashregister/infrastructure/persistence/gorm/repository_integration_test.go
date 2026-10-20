//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	cashRegisterGorm "github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/seed"
	"github.com/sigif/sigif-go/internal/modules/cashregister/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

func openDatabase(t *testing.T) *sharedDatabase.Database {
	t.Helper()
	dsn := os.Getenv("SIGIF_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SIGIF_TEST_DATABASE_DSN no está definido")
	}
	db, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.CashRegisterModel{}, &model.CashRegisterAssignmentModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(30)
	return &sharedDatabase.Database{DB: db}
}

// newCompany seeds the demo cash registers for a fresh company and removes
// them afterwards.
func newCompany(t *testing.T, db *sharedDatabase.Database) (uuid.UUID, map[string]uuid.UUID) {
	t.Helper()
	companyID := uuid.New()
	t.Cleanup(func() {
		db.DB.Where("company_id = ?", companyID).Delete(&model.CashRegisterAssignmentModel{})
		db.DB.Unscoped().Where("company_id = ?", companyID).Delete(&model.CashRegisterModel{})
	})
	if err := seed.Seed(context.Background(), db.DB, companyID); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var registers []model.CashRegisterModel
	db.DB.Where("company_id = ?", companyID).Find(&registers)
	ids := make(map[string]uuid.UUID, len(registers))
	for _, r := range registers {
		ids[r.Code] = r.ID
	}
	return companyID, ids
}

func newService(db *sharedDatabase.Database) (*service.CashRegisterService, repository.CashRegisterRepository) {
	repo := cashRegisterGorm.NewCashRegisterGormRepository(db)
	return service.NewCashRegisterService(repo, testutil.NewStubCashierDirectory(), clock.NewRealClock()), repo
}

func activeAssignments(t *testing.T, db *sharedDatabase.Database, column string, id uuid.UUID) int64 {
	t.Helper()
	var count int64
	db.DB.Model(&model.CashRegisterAssignmentModel{}).Where(column+" = ? AND status = 'active'", id).Count(&count)
	return count
}

func TestSeedIsIdempotent(t *testing.T) {
	db := openDatabase(t)
	companyID, ids := newCompany(t, db)

	if err := seed.Seed(context.Background(), db.DB, companyID); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	var registers []model.CashRegisterModel
	db.DB.Where("company_id = ?", companyID).Order("code").Find(&registers)
	if len(registers) != 3 || len(ids) != 3 {
		t.Fatalf("registers = %d, want 3", len(registers))
	}
	want := map[string]string{"CAJA-01": "active", "CAJA-02": "active", "CAJA-03": "inactive"}
	for _, r := range registers {
		if want[r.Code] != r.Status {
			t.Errorf("%s status = %s, want %s", r.Code, r.Status, want[r.Code])
		}
	}
}

func TestListRegistersWithAvailability(t *testing.T) {
	db := openDatabase(t)
	svc, repo := newService(db)
	ctx := context.Background()
	companyID, ids := newCompany(t, db)
	otherCompany, _ := newCompany(t, db)
	cashier := uuid.New()
	if _, err := svc.Assign(ctx, companyID, cashier, ids["CAJA-01"]); err != nil {
		t.Fatalf("assign: %v", err)
	}

	inUse, available := entity.AvailabilityInUse, entity.AvailabilityAvailable
	tests := []struct {
		name   string
		filter repository.ListFilter
		want   []string
	}{
		{"all", repository.ListFilter{}, []string{"CAJA-01", "CAJA-02"}},
		{"in use", repository.ListFilter{Availability: &inUse}, []string{"CAJA-01"}},
		{"available", repository.ListFilter{Availability: &available}, []string{"CAJA-02"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := repo.ListRegisters(ctx, companyID, tt.filter, 0, 10)
			if err != nil {
				t.Fatalf("ListRegisters() error = %v", err)
			}
			got := make([]string, len(items))
			for i, item := range items {
				got[i] = item.Register.Code
				if (item.Assignment != nil) != (item.Register.Code == "CAJA-01") {
					t.Errorf("%s assignment = %+v", item.Register.Code, item.Assignment)
				}
			}
			if total != int64(len(tt.want)) || len(got) != len(tt.want) || (len(got) > 0 && got[0] != tt.want[0]) {
				t.Errorf("codes = %v (total %d), want %v", got, total, tt.want)
			}
		})
	}

	t.Run("another company sees its registers available", func(t *testing.T) {
		items, _, _ := repo.ListRegisters(ctx, otherCompany, repository.ListFilter{Availability: &inUse}, 0, 10)
		if len(items) != 0 {
			t.Errorf("in_use registers of another company = %d, want 0", len(items))
		}
	})

	t.Run("deleted registers are not listed", func(t *testing.T) {
		db.DB.Where("cash_register_id = ?", ids["CAJA-02"]).Delete(&model.CashRegisterModel{})
		items, total, _ := repo.ListRegisters(ctx, companyID, repository.ListFilter{}, 0, 10)
		if total != 1 || len(items) != 1 || items[0].Register.Code != "CAJA-01" {
			t.Errorf("items = %d (total %d), want only CAJA-01", len(items), total)
		}
	})
}

func TestUniqueActiveIndexes(t *testing.T) {
	db := openDatabase(t)
	_, repo := newService(db)
	ctx := context.Background()
	companyID, ids := newCompany(t, db)
	clk := clock.NewRealClock()
	cashierA, cashierB := uuid.New(), uuid.New()

	if err := repo.CreateAssignment(ctx, entity.NewAssignment(clk, companyID, ids["CAJA-01"], cashierA)); err != nil {
		t.Fatalf("first assignment: %v", err)
	}

	t.Run("second active assignment of a register is rejected", func(t *testing.T) {
		err := repo.CreateAssignment(ctx, entity.NewAssignment(clk, companyID, ids["CAJA-01"], cashierB))
		if !errors.Is(err, service.ErrCashRegisterInUse) {
			t.Fatalf("error = %v, want ErrCashRegisterInUse", err)
		}
	})
	t.Run("second active assignment of a cashier is rejected", func(t *testing.T) {
		err := repo.CreateAssignment(ctx, entity.NewAssignment(clk, companyID, ids["CAJA-02"], cashierA))
		if !errors.Is(err, service.ErrCashierAlreadyAssigned) {
			t.Fatalf("error = %v, want ErrCashierAlreadyAssigned", err)
		}
	})
	t.Run("released assignments do not block", func(t *testing.T) {
		active, _ := repo.ActiveAssignmentByRegister(ctx, companyID, ids["CAJA-01"])
		active.Release(clk, cashierA)
		if released, err := repo.ReleaseAssignment(ctx, active); err != nil || !released {
			t.Fatalf("ReleaseAssignment() = %v, %v", released, err)
		}
		if err := repo.CreateAssignment(ctx, entity.NewAssignment(clk, companyID, ids["CAJA-01"], cashierB)); err != nil {
			t.Fatalf("assignment after release: %v", err)
		}
		var rows int64
		db.DB.Model(&model.CashRegisterAssignmentModel{}).Where("cash_register_id = ?", ids["CAJA-01"]).Count(&rows)
		if rows != 2 {
			t.Errorf("rows = %d, want 2: released assignments are kept as history", rows)
		}
	})
}

// TestConcurrentAssignmentOfSameRegister runs N cashiers taking the same cash
// register at the same time: exactly one wins and only one active assignment
// is stored (HU-10-02).
func TestConcurrentAssignmentOfSameRegister(t *testing.T) {
	db := openDatabase(t)
	svc, _ := newService(db)
	companyID, ids := newCompany(t, db)
	const cashiers = 20

	start := make(chan struct{})
	errs := make([]error, cashiers)
	var wg sync.WaitGroup
	for i := 0; i < cashiers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[i] = svc.Assign(context.Background(), companyID, uuid.New(), ids["CAJA-01"])
		}(i)
	}
	close(start)
	wg.Wait()

	won, inUse := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			won++
		case errors.Is(err, service.ErrCashRegisterInUse):
			inUse++
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}
	if won != 1 || inUse != cashiers-1 {
		t.Errorf("won = %d, in use = %d, want 1 and %d", won, inUse, cashiers-1)
	}
	if n := activeAssignments(t, db, "cash_register_id", ids["CAJA-01"]); n != 1 {
		t.Errorf("active assignments = %d, want 1", n)
	}
}

// TestConcurrentAssignmentOfSameCashier runs one cashier taking two cash
// registers at the same time: only one assignment is kept.
func TestConcurrentAssignmentOfSameCashier(t *testing.T) {
	db := openDatabase(t)
	svc, _ := newService(db)
	companyID, ids := newCompany(t, db)
	cashier := uuid.New()

	for round := 0; round < 10; round++ {
		start := make(chan struct{})
		errs := make([]error, 2)
		var wg sync.WaitGroup
		for i, code := range []string{"CAJA-01", "CAJA-02"} {
			wg.Add(1)
			go func(i int, register uuid.UUID) {
				defer wg.Done()
				<-start
				_, errs[i] = svc.Assign(context.Background(), companyID, cashier, register)
			}(i, ids[code])
		}
		close(start)
		wg.Wait()

		won := 0
		for _, err := range errs {
			if err == nil {
				won++
			} else if !errors.Is(err, service.ErrCashierAlreadyAssigned) && !errors.Is(err, service.ErrCashRegisterInUse) {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		if won != 1 {
			t.Fatalf("round %d: %d assignments succeeded, want 1 (%v)", round, won, errs)
		}
		if n := activeAssignments(t, db, "user_id", cashier); n != 1 {
			t.Fatalf("round %d: active assignments of the cashier = %d, want 1", round, n)
		}

		current, err := svc.Current(context.Background(), companyID, cashier)
		if err != nil {
			t.Fatalf("current: %v", err)
		}
		if _, err := svc.Release(context.Background(), service.ReleaseParams{CompanyID: companyID, ActorID: cashier, AssignmentID: current.Assignment.ID}); err != nil {
			t.Fatalf("release: %v", err)
		}
	}
}

func TestReleaseKeepsHistoryAndIsIdempotent(t *testing.T) {
	db := openDatabase(t)
	svc, _ := newService(db)
	ctx := context.Background()
	companyID, ids := newCompany(t, db)
	cashier, supervisor := uuid.New(), uuid.New()

	view, err := svc.Assign(ctx, companyID, cashier, ids["CAJA-01"])
	if err != nil {
		t.Fatalf("assign: %v", err)
	}
	params := service.ReleaseParams{CompanyID: companyID, ActorID: supervisor, AssignmentID: view.Assignment.ID, CanReleaseAny: true}

	var wg sync.WaitGroup
	results := make([]*service.AssignmentView, 5)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], _ = svc.Release(ctx, params)
		}(i)
	}
	wg.Wait()

	var row model.CashRegisterAssignmentModel
	db.DB.Where("assignment_id = ?", view.Assignment.ID).First(&row)
	if row.Status != "released" || row.ReleasedAt == nil || row.ReleasedBy == nil || *row.ReleasedBy != supervisor {
		t.Fatalf("row = %+v, want released by the supervisor", row)
	}
	for i, r := range results {
		// PostgreSQL keeps microseconds; the winning request returns its own
		// clock value with nanoseconds.
		if r == nil || r.Assignment.ReleasedAt == nil || r.Assignment.ReleasedAt.Sub(*row.ReleasedAt).Abs() >= time.Microsecond {
			t.Errorf("result %d = %+v, want the stored release at %v", i, r, row.ReleasedAt)
		}
	}

	if err := svc.EnsureCashierOwnsRegister(ctx, companyID, cashier, ids["CAJA-01"]); !errors.Is(err, service.ErrCashRegisterNotOwned) {
		t.Errorf("EnsureCashierOwnsRegister() after release = %v, want ErrCashRegisterNotOwned", err)
	}
	if _, err := svc.Assign(ctx, companyID, uuid.New(), ids["CAJA-01"]); err != nil {
		t.Errorf("assign after release: %v", err)
	}
	if _, err := svc.Assign(ctx, companyID, uuid.New(), ids["CAJA-03"]); !errors.Is(err, service.ErrCashRegisterNotFound) {
		t.Errorf("assign inactive register = %v, want ErrCashRegisterNotFound", err)
	}
}
