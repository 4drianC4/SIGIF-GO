package service_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	cashRegisterEntity "github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	cashRegisterService "github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
	"github.com/sigif/sigif-go/internal/modules/cashsession/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var now = time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC)

type fixture struct {
	svc       *service.CashSessionService
	sessions  *testutil.MemorySessionRepository
	registers *testutil.Registers
	company   uuid.UUID
}

func newFixture() *fixture {
	sessions := testutil.NewMemorySessionRepository()
	registers := testutil.NewRegisters()
	return &fixture{
		svc:       service.NewCashSessionService(sessions, registers.Access, clock.NewMockClock(now)),
		sessions:  sessions,
		registers: registers,
		company:   uuid.New(),
	}
}

func amount(s string) decimal.Decimal { return decimal.RequireFromString(s) }

// cashier returns a cashier holding a new active cash register.
func (f *fixture) cashier(t *testing.T, code string) (userID, registerID, assignmentID uuid.UUID) {
	t.Helper()
	userID = uuid.New()
	register := f.registers.Add(f.company, code, true)
	return userID, register.ID, f.registers.Assign(t, f.company, userID, register.ID)
}

func (f *fixture) open(userID uuid.UUID, opening string) (*service.SessionView, error) {
	return f.svc.Open(context.Background(), service.OpenParams{
		CompanyID: f.company, UserID: userID, OpeningAmount: amount(opening),
	})
}

func assertCode(t *testing.T, err error, code sharedErrors.ErrorCode, status int) {
	t.Helper()
	var appErr *sharedErrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != code || appErr.StatusCode != status {
		t.Fatalf("error = %v, want %s/%d", err, code, status)
	}
}

func TestOpenRecordsCashierRegisterAmountAndTime(t *testing.T) {
	f := newFixture()
	userID, registerID, assignmentID := f.cashier(t, "CAJA-01")

	view, err := f.svc.Open(context.Background(), service.OpenParams{
		CompanyID: f.company, UserID: userID, OpeningAmount: amount("150.50"), OpeningNotes: "  Billetes de 10  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := view.Session
	if s.UserID != userID || s.CompanyID != f.company || s.CashRegisterID != registerID || s.AssignmentID != assignmentID {
		t.Errorf("session ids = %+v", s)
	}
	if !s.OpeningAmount.Equal(amount("150.50")) || s.Status != entity.StatusOpen || !s.OpenedAt.Equal(now) || s.ClosedAt != nil {
		t.Errorf("session = %+v, want open at %s with 150.50", s, now)
	}
	if s.OpeningNotes != "Billetes de 10" {
		t.Errorf("notes = %q, want them trimmed", s.OpeningNotes)
	}
	if view.Register.ID != registerID || view.Register.Code != "CAJA-01" {
		t.Errorf("register = %+v", view.Register)
	}
	if f.sessions.OpenCount() != 1 {
		t.Errorf("open sessions = %d, want 1", f.sessions.OpenCount())
	}
}

func TestOpenAcceptsZeroAndRejectsInvalidAmounts(t *testing.T) {
	f := newFixture()
	userID, _, _ := f.cashier(t, "CAJA-01")
	if _, err := f.open(userID, "0"); err != nil {
		t.Fatalf("an opening amount of 0 must be valid: %v", err)
	}

	tests := []struct {
		name, amount, notes, field string
	}{
		{"negative", "-0.01", "", "opening_amount"},
		{"three decimals", "10.005", "", "opening_amount"},
		{"above the column", "10000000000", "", "opening_amount"},
		{"notes too long", "10", strings.Repeat("x", 201), "opening_notes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			userID, _, _ := f.cashier(t, "CAJA-01")
			_, err := f.svc.Open(context.Background(), service.OpenParams{
				CompanyID: f.company, UserID: userID, OpeningAmount: amount(tt.amount), OpeningNotes: tt.notes,
			})
			assertCode(t, err, sharedErrors.CodeValidation, 400)
			details, _ := err.(*sharedErrors.AppError).Details.(map[string]string)
			if _, ok := details[tt.field]; !ok {
				t.Errorf("details = %v, want %s", details, tt.field)
			}
			if f.sessions.OpenCount() != 0 {
				t.Error("an invalid opening was stored")
			}
		})
	}
}

func TestOpenRequiresAnActiveAssignment(t *testing.T) {
	f := newFixture()
	_, err := f.open(uuid.New(), "100")
	if !errors.Is(err, cashRegisterService.ErrNoActiveAssignment) {
		t.Fatalf("error = %v, want NO_ACTIVE_ASSIGNMENT", err)
	}
	if f.sessions.OpenCount() != 0 {
		t.Error("a session was opened without a cash register")
	}
}

func TestOpenRejectsAnInactiveCashRegister(t *testing.T) {
	f := newFixture()
	userID, registerID, _ := f.cashier(t, "CAJA-01")
	f.registers.Repo.Registers[registerID].Status = cashRegisterEntity.CashRegisterStatusInactive

	_, err := f.open(userID, "100")
	assertCode(t, err, service.CodeCashRegisterInactive, 409)
}

func TestOpenWithAnExplicitCashRegister(t *testing.T) {
	f := newFixture()
	userID, registerID, _ := f.cashier(t, "CAJA-01")
	_, otherRegisterID, _ := f.cashier(t, "CAJA-02")
	free := f.registers.Add(f.company, "CAJA-03", true)

	open := func(user, register uuid.UUID) error {
		_, err := f.svc.Open(context.Background(), service.OpenParams{
			CompanyID: f.company, UserID: user, CashRegisterID: &register, OpeningAmount: amount("50"),
		})
		return err
	}
	if err := open(userID, otherRegisterID); !errors.Is(err, cashRegisterService.ErrCashRegisterInUse) {
		t.Errorf("register of another cashier: error = %v, want CASH_REGISTER_IN_USE", err)
	}
	if err := open(userID, free.ID); !errors.Is(err, cashRegisterService.ErrCashRegisterNotOwned) {
		t.Errorf("free register: error = %v, want not owned", err)
	}
	if err := open(userID, uuid.New()); !errors.Is(err, cashRegisterService.ErrCashRegisterNotFound) {
		t.Errorf("unknown register: error = %v, want NOT_FOUND", err)
	}
	if err := open(userID, registerID); err != nil {
		t.Errorf("own register: %v", err)
	}
}

func TestOnlyOneOpenSessionPerCashierAndPerRegister(t *testing.T) {
	f := newFixture()
	userA, registerID, assignmentA := f.cashier(t, "CAJA-01")
	if _, err := f.open(userA, "100"); err != nil {
		t.Fatal(err)
	}

	_, err := f.open(userA, "200")
	if !errors.Is(err, service.ErrCashierHasOpenSession) {
		t.Fatalf("second opening: error = %v, want ErrCashierHasOpenSession", err)
	}

	// A releases the cash register without closing (RF-17 does not exist yet)
	// and B takes it: the register still has A's open session.
	f.registers.Release(t, f.company, userA, assignmentA)
	userB := uuid.New()
	f.registers.Assign(t, f.company, userB, registerID)
	_, err = f.open(userB, "100")
	if !errors.Is(err, service.ErrRegisterHasOpenSession) {
		t.Fatalf("register with an open session: error = %v, want ErrRegisterHasOpenSession", err)
	}
	assertCode(t, err, service.CodeCashSessionAlreadyOpen, 409)
}

func TestConcurrentOpeningsCreateASingleSession(t *testing.T) {
	f := newFixture()
	userID, _, _ := f.cashier(t, "CAJA-01")

	var wg sync.WaitGroup
	var mu sync.Mutex
	opened, conflicts := 0, 0
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.open(userID, "100")
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				opened++
			case errors.Is(err, service.ErrCashierHasOpenSession):
				conflicts++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
	if opened != 1 || conflicts != 19 || f.sessions.OpenCount() != 1 {
		t.Fatalf("opened = %d, conflicts = %d, stored = %d; want 1, 19, 1", opened, conflicts, f.sessions.OpenCount())
	}
}

func TestCurrent(t *testing.T) {
	f := newFixture()
	userID, _, _ := f.cashier(t, "CAJA-01")

	if _, err := f.svc.Current(context.Background(), f.company, userID); !errors.Is(err, service.ErrNoOpenSession) {
		t.Fatalf("without session: error = %v, want ErrNoOpenSession", err)
	}
	opened, err := f.open(userID, "80")
	if err != nil {
		t.Fatal(err)
	}
	view, err := f.svc.Current(context.Background(), f.company, userID)
	if err != nil || view.Session.ID != opened.Session.ID || view.Register.Code != "CAJA-01" {
		t.Fatalf("current = %+v, %v", view, err)
	}
	if _, err := f.svc.Current(context.Background(), uuid.New(), userID); !errors.Is(err, service.ErrNoOpenSession) {
		t.Errorf("another company: error = %v, want ErrNoOpenSession", err)
	}

	f.sessions.Close(opened.Session.ID)
	if _, err := f.svc.Current(context.Background(), f.company, userID); !errors.Is(err, service.ErrNoOpenSession) {
		t.Errorf("closed session: error = %v, want ErrNoOpenSession", err)
	}
}

func TestRequireOpenSession(t *testing.T) {
	f := newFixture()
	userID, _, assignmentID := f.cashier(t, "CAJA-01")

	_, err := f.svc.RequireOpenSession(context.Background(), f.company, userID)
	assertCode(t, err, service.CodeCashSessionRequired, 409)

	opened, err := f.open(userID, "100")
	if err != nil {
		t.Fatal(err)
	}
	session, err := f.svc.RequireOpenSession(context.Background(), f.company, userID)
	if err != nil || session.ID != opened.Session.ID {
		t.Fatalf("with an open session: %+v, %v", session, err)
	}

	// The session stays open but the cash register is no longer the cashier's.
	f.registers.Release(t, f.company, userID, assignmentID)
	if _, err := f.svc.RequireOpenSession(context.Background(), f.company, userID); !errors.Is(err, cashRegisterService.ErrCashRegisterNotOwned) {
		t.Errorf("released register: error = %v, want not owned", err)
	}
}
