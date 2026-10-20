//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	cashRegisterService "github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	cashRegisterGorm "github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/gorm"
	cashRegisterModel "github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/model"
	cashRegisterSeed "github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/seed"
	cashRegisterTestutil "github.com/sigif/sigif-go/internal/modules/cashregister/testutil"
	"github.com/sigif/sigif-go/internal/modules/cashsession/application/adapter"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
	cashSessionGorm "github.com/sigif/sigif-go/internal/modules/cashsession/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/cashsession/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type env struct {
	db        *sharedDatabase.Database
	registers *cashRegisterService.CashRegisterService
	sessions  *service.CashSessionService
	company   uuid.UUID
	codes     map[string]uuid.UUID
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dsn := os.Getenv("SIGIF_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SIGIF_TEST_DATABASE_DSN no está definido")
	}
	gdb, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := gdb.AutoMigrate(&cashRegisterModel.CashRegisterModel{}, &cashRegisterModel.CashRegisterAssignmentModel{}, &model.CashSessionModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	sqlDB, _ := gdb.DB()
	sqlDB.SetMaxOpenConns(30)
	db := &sharedDatabase.Database{DB: gdb}

	companyID := uuid.New()
	t.Cleanup(func() {
		gdb.Where("company_id = ?", companyID).Delete(&model.CashSessionModel{})
		gdb.Where("company_id = ?", companyID).Delete(&cashRegisterModel.CashRegisterAssignmentModel{})
		gdb.Unscoped().Where("company_id = ?", companyID).Delete(&cashRegisterModel.CashRegisterModel{})
	})
	if err := cashRegisterSeed.Seed(context.Background(), gdb, companyID); err != nil {
		t.Fatalf("seed cash registers: %v", err)
	}
	var registers []cashRegisterModel.CashRegisterModel
	gdb.Where("company_id = ?", companyID).Find(&registers)
	codes := make(map[string]uuid.UUID, len(registers))
	for _, r := range registers {
		codes[r.Code] = r.ID
	}

	registerService := cashRegisterService.NewCashRegisterService(
		cashRegisterGorm.NewCashRegisterGormRepository(db), cashRegisterTestutil.NewStubCashierDirectory(), clock.NewRealClock())
	sessionService := service.NewCashSessionService(
		cashSessionGorm.NewCashSessionGormRepository(db), adapter.NewRegisterAccessAdapter(registerService), clock.NewRealClock())
	return &env{db: db, registers: registerService, sessions: sessionService, company: companyID, codes: codes}
}

func (e *env) cashier(t *testing.T, code string) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	if _, err := e.registers.Assign(context.Background(), e.company, userID, e.codes[code]); err != nil {
		t.Fatalf("assign %s: %v", code, err)
	}
	return userID
}

func (e *env) open(userID uuid.UUID, amount string) (*service.SessionView, error) {
	return e.sessions.Open(context.Background(), service.OpenParams{
		CompanyID: e.company, UserID: userID, OpeningAmount: decimal.RequireFromString(amount), OpeningNotes: "Fondo inicial",
	})
}

func (e *env) openRows(t *testing.T) int64 {
	t.Helper()
	var count int64
	e.db.DB.Model(&model.CashSessionModel{}).Where("company_id = ? AND status = 'open'", e.company).Count(&count)
	return count
}

func TestOpenAndReadSessionWithPostgres(t *testing.T) {
	e := newEnv(t)
	userID := e.cashier(t, "CAJA-01")

	opened, err := e.open(userID, "0")
	if err != nil {
		t.Fatal(err)
	}
	var row model.CashSessionModel
	if err := e.db.DB.First(&row, "cash_session_id = ?", opened.Session.ID).Error; err != nil {
		t.Fatal(err)
	}
	if row.Status != "open" || !row.OpeningAmount.IsZero() || row.CashRegisterID != e.codes["CAJA-01"] ||
		row.UserID != userID || row.OpeningNotes == nil || *row.OpeningNotes != "Fondo inicial" || row.ClosedAt != nil {
		t.Fatalf("stored row = %+v", row)
	}

	current, err := e.sessions.Current(context.Background(), e.company, userID)
	if err != nil || current.Session.ID != opened.Session.ID || current.Register.Code != "CAJA-01" ||
		current.Session.Status != entity.StatusOpen {
		t.Fatalf("current = %+v, %v", current, err)
	}
	if _, err := e.sessions.RequireOpenSession(context.Background(), e.company, userID); err != nil {
		t.Fatalf("require open session: %v", err)
	}
}

func TestDatabaseRejectsInvalidRows(t *testing.T) {
	e := newEnv(t)
	userID := e.cashier(t, "CAJA-01")
	opened, err := e.open(userID, "10")
	if err != nil {
		t.Fatal(err)
	}
	if err := e.db.DB.Model(&model.CashSessionModel{}).Where("cash_session_id = ?", opened.Session.ID).
		Update("opening_amount", decimal.NewFromInt(-1)).Error; err == nil {
		t.Error("a negative opening amount was accepted by the database")
	}
	if err := e.db.DB.Model(&model.CashSessionModel{}).Where("cash_session_id = ?", opened.Session.ID).
		Update("status", "abierta").Error; err == nil {
		t.Error("an unknown status was accepted by the database")
	}
}

// The indexes reject the duplicate even when every request passes the checks
// at the same time.
func TestConcurrentOpeningsWithPostgres(t *testing.T) {
	e := newEnv(t)
	userID := e.cashier(t, "CAJA-01")

	var wg sync.WaitGroup
	var mu sync.Mutex
	opened, conflicts := 0, 0
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := e.open(userID, "100")
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				opened++
			case errors.Is(err, service.ErrCashierHasOpenSession), errors.Is(err, service.ErrRegisterHasOpenSession):
				conflicts++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
	if opened != 1 || conflicts != 19 || e.openRows(t) != 1 {
		t.Fatalf("opened = %d, conflicts = %d, rows = %d; want 1, 19, 1", opened, conflicts, e.openRows(t))
	}
}

func TestRegisterWithOpenSessionCannotOpenAnother(t *testing.T) {
	e := newEnv(t)
	userA := e.cashier(t, "CAJA-01")
	if _, err := e.open(userA, "100"); err != nil {
		t.Fatal(err)
	}
	view, err := e.registers.Current(context.Background(), e.company, userA)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.registers.Release(context.Background(), cashRegisterService.ReleaseParams{
		CompanyID: e.company, ActorID: userA, AssignmentID: view.Assignment.ID,
	}); err != nil {
		t.Fatal(err)
	}
	userB := e.cashier(t, "CAJA-01")
	if _, err := e.open(userB, "100"); !errors.Is(err, service.ErrRegisterHasOpenSession) {
		t.Fatalf("error = %v, want ErrRegisterHasOpenSession", err)
	}
	if e.openRows(t) != 1 {
		t.Errorf("open rows = %d, want 1", e.openRows(t))
	}
}
