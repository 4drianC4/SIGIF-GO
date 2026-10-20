package handler_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashsession/application/adapter"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// salesPath stands in for a cash transaction of another module (RF-11): it is
// guarded exactly as those routes must be.
const salesPath = "/api/v1/test-sales"

func (s *testServer) mountSales(t *testing.T) *int {
	t.Helper()
	runs := 0
	s.app.Post(salesPath,
		middleware.RequireOpenCashSession(adapter.NewSessionCheckerAdapter(s.service)),
		func(c *fiber.Ctx) error {
			if _, ok := middleware.CashSessionIDFromContext(c.UserContext()); !ok {
				t.Error("the guarded handler did not receive the cash session")
			}
			runs++
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true})
		})
	return &runs
}

func TestCashTransactionsRequireAnOpenSession(t *testing.T) {
	s := newTestServer(t)
	runs := s.mountSales(t)
	userID, token, _ := s.cashier(t, "CAJA-01")

	status, resp := s.send(t, fiber.MethodPost, salesPath, token, map[string]any{})
	expectError(t, status, resp, 409, "CASH_SESSION_REQUIRED")
	if *runs != 0 {
		t.Fatal("a transaction ran without an open session")
	}

	if status, resp, _ := s.open(t, token, map[string]any{"opening_amount": 100}); status != fiber.StatusCreated {
		t.Fatalf("open: %d %s", status, resp.raw)
	}
	status, resp = s.send(t, fiber.MethodPost, salesPath, token, map[string]any{})
	if status != fiber.StatusCreated || *runs != 1 {
		t.Fatalf("with an open session: %d %s, runs = %d", status, resp.raw, *runs)
	}

	// Releasing the cash register (HU-10) while the session is open stops the
	// transactions: the register is no longer the cashier's.
	view, err := s.registers.Service.Current(t.Context(), uuid.MustParse(companyA), userID)
	if err != nil {
		t.Fatal(err)
	}
	s.registers.Release(t, uuid.MustParse(companyA), userID, view.Assignment.ID)
	status, resp = s.send(t, fiber.MethodPost, salesPath, token, map[string]any{})
	expectError(t, status, resp, 403, "FORBIDDEN")
	if *runs != 1 {
		t.Fatal("a transaction ran on a released cash register")
	}
}

func TestCashTransactionOfAnotherCashierIsBlocked(t *testing.T) {
	s := newTestServer(t)
	runs := s.mountSales(t)
	_, tokenA, _ := s.cashier(t, "CAJA-01")
	_, tokenB, _ := s.cashier(t, "CAJA-02")
	if status, resp, _ := s.open(t, tokenA, map[string]any{"opening_amount": 100}); status != fiber.StatusCreated {
		t.Fatalf("open: %d %s", status, resp.raw)
	}

	status, resp := s.send(t, fiber.MethodPost, salesPath, tokenB, map[string]any{})
	expectError(t, status, resp, 409, "CASH_SESSION_REQUIRED")
	if *runs != 0 {
		t.Fatal("the session of cashier A let cashier B transact")
	}
}
